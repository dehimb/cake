package server

import (
	"bytes"
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

func (storeHandler *MockStoreHandler) IsTokenValid(token string) bool {
	return true
}

func (storeHandler *MockStoreHandler) CreateUser(user *store.User) error {
	return nil
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

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name string
		body string
		code int
	}{
		{
			name: "Valid request",
			body: `{"token": "tkn", "id": 10, "balance": 100}`,
			code: http.StatusOK,
		},
		{
			name: "Invalid request",
			body: `malformed json`,
			code: http.StatusBadRequest,
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
			assert.Equal(t, testCase.code, rec.Code)
		})
	}
}
