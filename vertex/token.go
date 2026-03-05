package vertex

import (
	"fmt"

	"golang.org/x/oauth2"
)

// oauth2TokenSource adapts golang.org/x/oauth2.TokenSource to our simple Token() string interface.
type oauth2TokenSource struct {
	ts oauth2.TokenSource
}

func (s *oauth2TokenSource) Token() (string, error) {
	t, err := s.ts.Token()
	if err != nil {
		return "", fmt.Errorf("refreshing token: %w", err)
	}
	return t.AccessToken, nil
}
