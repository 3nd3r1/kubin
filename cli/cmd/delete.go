package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/tokens"
	"github.com/3nd3r1/kubin/cli/pkg/upload"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	// Token overrides the delete token stored by kubin create.
	Token string
}

var deleteOpts deleteOptions

var deleteCmd = &cobra.Command{
	Use:   "delete <url>",
	Short: "Delete a snapshot from its kubin server",
	Long: `Delete a snapshot using the delete token saved by kubin create, or a
token passed with --token.`,
	Example: `  kubin delete https://kubin.example.com/s/7fK2q
  kubin delete --token <token> https://kubin.example.com/s/7fK2q`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(cmd.Context(), args[0], deleteOpts)
	},
}

func init() {
	deleteCmd.Flags().StringVar(&deleteOpts.Token, "token", "",
		"delete token for the snapshot (default: read from the local token file)")
}

func runDelete(ctx context.Context, rawURL string, opts deleteOptions) error {
	serverBase, id, canonicalURL, err := splitSnapshotURL(rawURL)
	if err != nil {
		return err
	}

	token := opts.Token
	if token == "" {
		stored, ok, err := tokens.Get(canonicalURL)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no delete token stored for %s; pass --token to supply one", canonicalURL)
		}
		token = stored
	}

	err = upload.New(serverBase, "kubin/"+version).Delete(ctx, id, token)
	switch {
	case err == nil:
		if delErr := tokens.Delete(canonicalURL); delErr != nil {
			log.Warn("Could not remove stored delete token", "error", delErr)
		}
		log.Info("Deleted " + canonicalURL)
		return nil
	case errors.Is(err, upload.ErrNotFound):
		if delErr := tokens.Delete(canonicalURL); delErr != nil {
			log.Warn("Could not remove stored delete token", "error", delErr)
		}
		log.Warn(fmt.Sprintf("Snapshot %s not found; it may have already expired or been deleted", canonicalURL))
		return nil
	case errors.Is(err, upload.ErrForbidden):
		return fmt.Errorf("could not delete %s: invalid delete token", canonicalURL)
	default:
		return err
	}
}

// splitSnapshotURL splits a snapshot link at the last "/s/" into the server
// base URL, the snapshot ID, and the canonical snapshot URL.
func splitSnapshotURL(rawURL string) (serverBase, id, canonicalURL string, err error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", "", "", fmt.Errorf("snapshot URL %q must be a link like https://kubin.example.com/s/<id>", rawURL)
	}
	cleanPath := strings.TrimRight(u.Path, "/")
	idx := strings.LastIndex(cleanPath, "/s/")
	if idx < 0 {
		return "", "", "", fmt.Errorf("snapshot URL %q must contain /s/<id>", rawURL)
	}
	prefix := cleanPath[:idx]
	id = cleanPath[idx+len("/s/"):]
	if id == "" || strings.Contains(id, "/") {
		return "", "", "", fmt.Errorf("snapshot URL %q must end with /s/<id>", rawURL)
	}
	serverBase = u.Scheme + "://" + u.Host + prefix
	canonicalURL = serverBase + "/s/" + id
	return serverBase, id, canonicalURL, nil
}
