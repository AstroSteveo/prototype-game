//go:build !ws

package ws

import (
	"net/http"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/sim"
)

// Register installs a placeholder handler when the websocket build tag is not enabled.
// Build with `-tags ws` to enable the real websocket server.
func Register(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "websocket transport not built (use -tags ws)", http.StatusNotImplemented)
	})
}
