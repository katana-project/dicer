package api

import (
	"net/http"

	"github.com/charmbracelet/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("request received", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}
