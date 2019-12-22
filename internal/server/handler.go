package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type handler struct {
	router       *mux.Router
	storeHandler store.StoreHandler
	logger       *logrus.Logger
}

func (h *handler) initRouter(m MiddlewareDispatcher) {
	// Provide all middlewares from one method
	h.router.Use(m.populate()...)

	h.router.HandleFunc("/user", h.userGet).Methods("GET")
	h.router.HandleFunc("/user", h.userPost).Methods("POST")
	h.router.PathPrefix("/").HandlerFunc(h.defaultHandler)
}

func (h *handler) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// Create new user
func (h *handler) userPost(w http.ResponseWriter, r *http.Request) {
	var u store.User
	err := h.parseRequestBody(r, &u)
	if err != nil {
		h.logger.Info("Bad request for user create: ", err)
		h.sendResponse(w, http.StatusBadRequest, &ErrorResponse{Error: "Invalid request"})
		return
	}
	err = h.storeHandler.CreateUser(&u)
	if err != nil {
		var validationError *store.ValidationError
		var internalError *store.InternalError
		switch {
		case errors.As(err, &validationError):
			h.sendResponse(w, http.StatusBadRequest, &ErrorResponse{Error: err.Error()})
			return
		case errors.As(err, &internalError):
			h.logger.Error("Error when processing request: ", err)
			h.sendResponse(w, http.StatusInternalServerError, &ErrorResponse{Error: "Internal server error"})
			return
		default:
			h.logger.Warn("Unhadled error: ", err)
			h.sendResponse(w, http.StatusInternalServerError, &ErrorResponse{Error: "Internal server error"})
			return
		}
	}
	h.sendResponse(w, http.StatusOK, &UserCreateResponse{})
}
func (h *handler) userGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *handler) parseRequestBody(r *http.Request, dst interface{}) error {
	// TODO move content type check to middleware
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("Invalid content type")
	}
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&dst)
	if err != nil || dec.More() {
		return err
	}
	return nil
}

func (h handler) sendResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Error when tryibg marshal response: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
	w.Write(json)
}
