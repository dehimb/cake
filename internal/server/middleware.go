package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
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
	logger       *logrus.Logger
	storeHandler store.StoreHandler
}

type MiddlewareDispatcher interface {
	populate() []mux.MiddlewareFunc
}

// Validate token from client request
func IsTokenValid(token string) bool {
	return len(token) > 0
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
			if r.Method == "POST" || r.Method == "PUT" {
				body, _ := ioutil.ReadAll(r.Body)
				m.logger.Info("Body: ", string(body))
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
			next.ServeHTTP(w, r)
			m.logger.Infof("<-  %s %s %s", time.Since(start), r.Method, r.URL)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TODO add maxBytesReader middleware

// Validate tokens for every client request
func (m *middleware) checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if token := r.URL.Query().Get("token"); IsTokenValid(token) {
				next.ServeHTTP(w, r)
				return
			}
			sendErrorResponse(w, "Invalid token", http.StatusBadRequest)
		case "POST", "PUT":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				m.logger.Error("Failed to read request body: ", err)
				sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			var data map[string]interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				sendErrorResponse(w, "Invalid token", http.StatusBadRequest)
				return
			}

			if token, _ := data["token"].(string); IsTokenValid(token) {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				next.ServeHTTP(w, r)
				return
			}
			sendErrorResponse(w, "Invalid token", http.StatusBadRequest)
		default:
			sendErrorResponse(w, "Not found", http.StatusNotFound)
		}
	})
}

// TODO add validator for content type

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
