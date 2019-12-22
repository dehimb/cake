package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var middlewareTestHandler *handler

type MiddlewareMockStoreHandler struct {
}

func (storeHandler *MiddlewareMockStoreHandler) GetUser(userID uint64) (*store.User, *store.Statistic, error) {
	return nil, nil, nil
}

func (storeHandler *MiddlewareMockStoreHandler) CreateUser(user *store.User) error {
	return nil
}

func (storeHandler *MiddlewareMockStoreHandler) CreateDeposit(d *store.Deposit) (float32, error) {
	return 0, nil
}

func init() {
	middlewareTestHandler = &handler{
		router:       mux.NewRouter(),
		storeHandler: &MiddlewareMockStoreHandler{},
		logger:       logrus.New(),
	}

	m := &middleware{
		logger:       middlewareTestHandler.logger,
		storeHandler: middlewareTestHandler.storeHandler,
	}
	middlewareTestHandler.initRouter(m)
}

func TestCors(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	middlewareTestHandler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCheckToken(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		url          string
		data         io.Reader
		expectedCode int
	}{
		{
			name:         "Valid GET",
			method:       "GET",
			url:          "/ping?token=tkn",
			data:         nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid GET",
			method:       "GET",
			url:          "/ping",
			data:         nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Valid POST",
			method:       "POST",
			url:          "/ping",
			data:         bytes.NewBuffer([]byte(`{"token":"tkn"}`)),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid POST",
			method:       "POST",
			url:          "/ping",
			data:         bytes.NewBuffer([]byte(`{"data":"tkn"}`)),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Valid PUT",
			method:       "PUT",
			url:          "/ping",
			data:         bytes.NewBuffer([]byte(`{"token":"tkn"}`)),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid PUT",
			method:       "PUT",
			url:          "/ping",
			data:         bytes.NewBuffer([]byte(`{"data":"tkn"}`)),
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(testCase.method, testCase.url, testCase.data)
			if testCase.method == "POST" || testCase.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}
			middlewareTestHandler.ServeHTTP(rec, req)
			bodyBytes, _ := ioutil.ReadAll(rec.Body)
			t.Log(string(bodyBytes))
			assert.Equal(t, testCase.expectedCode, rec.Code)
		})
	}
}
