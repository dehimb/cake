package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

	h.router.HandleFunc("/user", h.userPost).Methods("POST")
	h.router.HandleFunc("/user", h.userGet).Methods("GET")
	h.router.HandleFunc("/user/deposit", h.depositPost).Methods("POST")
	h.router.HandleFunc("/ping", h.ping).Methods("GET", "POST", "PUT")
	h.router.PathPrefix("/").HandlerFunc(h.defaultHandler)
}

func (h *handler) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *handler) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *handler) depositPost(w http.ResponseWriter, r *http.Request) {
	var d store.Deposit
	err := h.parseRequestBody(r, &d)
	if err != nil {
		h.logger.Info("Bad request for deposit: ", err)
		sendErrorResponse(w, "Invalid request", http.StatusBadRequest)
		return
	}
	balance, err := h.storeHandler.CreateDeposit(&d)
	if err != nil {
		h.processError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, &DepositResponse{
		Balance: balance,
		Errror:  "",
	})
}

func (h *handler) userGet(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil || userID == 0 {
		sendErrorResponse(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	user, statistic, err := h.storeHandler.GetUser(userID)
	if err != nil {
		h.processError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, &UserResponse{
		UserID:        user.ID,
		Balance:       user.Balance,
		DepositeCount: statistic.DepositeCount,
		DepositSum:    statistic.DepositSum,
		BetCount:      statistic.BetCount,
		BetSum:        statistic.BetSum,
		WinCount:      statistic.WinCount,
		WinSum:        statistic.WinSum,
	})
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
		h.processError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, &UserCreateResponse{})
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

func (h *handler) processError(w http.ResponseWriter, err error) {
	var validationError *store.ValidationError
	var internalError *store.InternalError
	var notFoundError *store.NotFoundError
	var transactionError *store.TransactionError
	switch {
	case errors.As(err, &validationError):
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	case errors.As(err, &internalError):
		h.logger.Error("Error when processing request: ", err)
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	case errors.As(err, &notFoundError):
		sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	case errors.As(err, &transactionError):
		sendErrorResponse(w, "Transaction error", http.StatusBadRequest)
		return
	default:
		h.logger.Warn("Unhadled error: ", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	json, _ := json.Marshal(&ErrorResponse{Error: message})
	w.Write(json)

}
