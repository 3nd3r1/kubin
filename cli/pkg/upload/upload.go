// Package upload sends a snapshot tarball to a kubin server.
package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Result is what the server answers to an upload.
type Result struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	ExpiresAt   time.Time `json:"expires_at"`
	DeleteToken string    `json:"delete_token"`
}

var (
	// ErrNotFound is returned when the server reports the snapshot does not exist.
	ErrNotFound = errors.New("snapshot not found")
	// ErrForbidden is returned when the server rejects the delete token.
	ErrForbidden = errors.New("invalid delete token")
)

type Client struct {
	// Server is the base URL, like https://kubin.example.com.
	Server    string
	HTTP      *http.Client
	UserAgent string
}

func New(server, userAgent string) *Client {
	return &Client{
		Server:    server,
		HTTP:      &http.Client{Timeout: 5 * time.Minute},
		UserAgent: userAgent,
	}
}

// File uploads the tarball at path. When expire is non-empty, it is sent
// as the expires_in query parameter.
func (c *Client) File(ctx context.Context, path, expire string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return Result{}, err
	}
	return c.Upload(ctx, f, st.Size(), expire)
}

// Upload sends size bytes from r as a snapshot. When expire is non-empty,
// it is sent as the expires_in query parameter.
func (c *Client) Upload(ctx context.Context, r io.Reader, size int64, expire string) (Result, error) {
	endpoint, err := c.endpoint()
	if err != nil {
		return Result{}, err
	}
	if expire != "" {
		endpoint += "?" + url.Values{"expires_in": {expire}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, r)
	if err != nil {
		return Result{}, err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return Result{}, fmt.Errorf("can't reach the kubin server at %s: %w", c.Server, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, fmt.Errorf("read the server response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("the kubin server at %s refused the snapshot: %s", c.Server, serverMessage(resp, body))
	}
	var res Result
	if err := json.Unmarshal(body, &res); err != nil || res.URL == "" {
		return Result{}, fmt.Errorf("the kubin server at %s sent an unexpected answer (%s); is it a kubin server?", c.Server, resp.Status)
	}
	return res, nil
}

// Delete removes snapshot id using the bearer delete token.
func (c *Client) Delete(ctx context.Context, id, token string) error {
	endpoint, err := c.endpoint()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint+"/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return fmt.Errorf("can't reach the kubin server at %s: %w", c.Server, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read the server response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, serverMessage(resp, body))
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, serverMessage(resp, body))
	default:
		return fmt.Errorf("the kubin server at %s refused to delete the snapshot: %s", c.Server, serverMessage(resp, body))
	}
}

func (c *Client) endpoint() (string, error) {
	u, err := url.Parse(c.Server)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("server %q must be a URL like https://kubin.example.com", c.Server)
	}
	return strings.TrimRight(c.Server, "/") + "/api/v1/snapshots", nil
}

// serverMessage is the server's error message, or the status if it sent
// none.
func serverMessage(resp *http.Response, body []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		return fmt.Sprintf("%s (%s)", e.Error, resp.Status)
	}
	return resp.Status
}
