package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testHandler *handler

type MockStoreHandler struct {
}

func (storeHandler *MockStoreHandler) CreateUser(user *store.User) error {
	return nil
}

func (storeHandler *MockStoreHandler) CreateDeposit(d *store.Deposit) (float32, error) {
	return 0, nil
}

func (storeHandler *MockStoreHandler) GetUser(userID uint64) (*store.User, *store.Statistic, error) {
	if userID == 1 {
		return &store.User{}, &store.Statistic{}, nil
	}
	if userID == 2 {
		return nil, nil, errors.New("Unknown error")
	}
	return nil, nil, &store.NotFoundError{}
}

type MockMiddlware struct {
}

func (middleware *MockMiddlware) populate() []mux.MiddlewareFunc {
	return make([]mux.MiddlewareFunc, 0)
}

func init() {
	testHandler = &handler{
		router:       mux.NewRouter(),
		storeHandler: &MockStoreHandler{},
		logger:       logrus.New(),
	}
	testHandler.initRouter(&MockMiddlware{})
}

func TestUserPost(t *testing.T) {
	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "Valid request",
			body:         `{"token": "tkn", "id": 10, "balance": 100}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid request",
			body:         `malformed json`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer([]byte(testCase.body)))
			req.Header.Set("Content-Type", "application/json")
			testHandler.ServeHTTP(rec, req)
			bodyBytes, _ := ioutil.ReadAll(rec.Body)
			t.Log(string(bodyBytes))
			assert.Equal(t, testCase.expectedCode, rec.Code)
		})
	}

}
func TestUserGet(t *testing.T) {
	testCases := []struct {
		name         string
		userID       uint64
		expectedCode int
	}{
		{
			name:         "Invalid user id",
			userID:       0,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Valid request",
			userID:       1,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Unknown error",
			userID:       2,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "User not found",
			userID:       3332,
			expectedCode: http.StatusNotFound,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			// _, _, err := testHandler.storeHandler.GetUser(testCase.userID)
			t.Logf("ERR: %T", testHandler.storeHandler)
			req, _ := http.NewRequest("GET", fmt.Sprintf("/user?token=tkn&id=%d", testCase.userID), nil)
			testHandler.ServeHTTP(rec, req)
			bodyBytes, _ := ioutil.ReadAll(rec.Body)
			t.Log(string(bodyBytes))
			assert.Equal(t, testCase.expectedCode, rec.Code)

		})
	}
}

func TestDepositPost(t *testing.T) {
	t.Error("ERR")
}
