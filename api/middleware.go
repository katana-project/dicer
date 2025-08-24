package api

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)

		log.Debug("request received", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr, "time", time.Since(t).String())
	})
}
