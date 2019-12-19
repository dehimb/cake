package server

import (
	"net/http"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type handler struct {
	router *mux.Router
	store  *store.Store
	logger *logrus.Logger
}

func (h *handler) initRouter() {
	m := &middleware{
		logger: h.logger,
		store:  h.store,
	}
	// Provide all middlewares from one method
	h.router.Use(m.populate()...)

	h.router.HandleFunc("/user", h.userGet).Methods("GET")
	h.router.HandleFunc("/user", h.userPost).Methods("POST")
	h.router.PathPrefix("/").HandlerFunc(h.defaultHandler)
}

func (h *handler) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *handler) userPost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func (h *handler) userGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
