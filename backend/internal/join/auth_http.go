package join

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// HTTPAuth is an AuthService implementation that calls a gateway /validate endpoint.
type HTTPAuth struct {
	Client  *http.Client
	BaseURL string // e.g., http://localhost:8080
}

func NewHTTPAuth(base string) *HTTPAuth {
	return &HTTPAuth{Client: &http.Client{Timeout: 3 * time.Second}, BaseURL: base}
}

func (h *HTTPAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	u := h.BaseURL + "/validate?token=" + url.QueryEscape(token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := h.Client.Do(req)
	if err != nil {
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", false
	}
	var out struct {
		PlayerID string `json:"player_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", false
	}
	return out.PlayerID, out.Name, true
}
