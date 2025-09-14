//go:build !ws

package ws

import (
	"net/http"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/state"
)

// Register installs a placeholder handler when the websocket build tag is not enabled.
// Build with `-tags ws` to enable the real websocket server.
func Register(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine) {
	RegisterWithStore(mux, path, auth, eng, nil)
}

// RegisterWithStore is a placeholder when ws is disabled.
func RegisterWithStore(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine, _ state.Store) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "websocket transport not built (use -tags ws)", http.StatusNotImplemented)
	})
}
