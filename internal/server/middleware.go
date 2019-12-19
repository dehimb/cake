package server

import (
	"net/http"
	"time"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Middlewares functions works like interceptors for every http request.
// Methods provides ability to stop or propagate request to the chain.
type middleware struct {
	logger *logrus.Logger
	store  *store.Store
}

// This method used to check preflight requests
func (m *middleware) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Used for loggin request method, url and execution time.
// Log only when log level set to logrus.InfoLevel or higher.
func (m *middleware) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.logger.Level >= logrus.InfoLevel {
			start := time.Now()
			m.logger.Infof("-> %s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
			m.logger.Infof("<-  %s %s %s", time.Since(start), r.Method, r.URL)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Validate tokens for every client request
func (m *middleware) checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if token := r.URL.Query().Get("token"); m.store.IsTokenValid(token) {
				next.ServeHTTP(w, r)
				return
			}
			w.Write([]byte("Invalid token"))
			w.WriteHeader(http.StatusForbidden)
		case "POST", "PUT":
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// Method used for providing all middlewares at one place
// Declare all midlwares and add them to return array
func (m *middleware) populate() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{
		m.logRequest,
		m.cors,
		handlers.CORS(handlers.AllowedOrigins([]string{"*"})),
		m.checkToken,
	}
}
