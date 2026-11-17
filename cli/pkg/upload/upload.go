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
	ID  string `json:"id"`
	URL string `json:"url"`
}

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

// File uploads the tarball at path.
func (c *Client) File(ctx context.Context, path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return Result{}, err
	}
	return c.Upload(ctx, f, st.Size())
}

// Upload sends size bytes from r as a snapshot.
func (c *Client) Upload(ctx context.Context, r io.Reader, size int64) (Result, error) {
	endpoint, err := c.endpoint()
	if err != nil {
		return Result{}, err
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
