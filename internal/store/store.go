// Currently this package operates local database and in-memory cache
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Store struct {
	logger         *logrus.Logger
	db             *sql.DB
	users          map[uint64]*User
	pendingActions PendingActions
}

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return e.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type InternalError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

func New(ctx context.Context, logger *logrus.Logger) *Store {
	s := &Store{logger: logger}
	s.init(ctx)
	return s
}

func (s *Store) init(ctx context.Context) {
	var err error
	s.db, err = sql.Open("sqlite3", "cake.db")
	if err != nil {
		s.logger.Fatal("Can't open database: ", err)
	}
	_, err = s.db.Exec(CreateTables)
	if err != nil {
		s.logger.Fatal("Can't create tables: ", err)
	}
	_, err = s.db.Exec(CreateIndexes)
	if err != nil {
		s.logger.Fatal("Can't create indexes: ", err)
	}
	go func() {
		// Listen for context done signal and close connection
		<-ctx.Done()
		err := s.db.Close()
		if err != nil {
			s.logger.Error("Can't close database: ", err)
			return
		}
		s.logger.Info("Database connection closed")
	}()
}

// Validate token from client request
func (s *Store) IsTokenValid(token string) bool {
	return len(token) > 0
}

// Create new user or return error
func (s *Store) createUser(user *User) error {
	// check, is user already exists
	if _, ok := s.users[user.ID]; ok {
		return &ValidationError{errors.New("User already exists")}
	}
	// check balance
	if user.Balance <= 0 {
		return &ValidationError{errors.New("User balance may be greater than 0")}
	}
	stmt, err := s.db.Prepare("INSERT INTO users(id, name) values(?, ?)")
	if err != nil {
		return &InternalError{Message: "Error when creating db statement", Err: err}
	}
	if _, err = stmt.Exec(user.ID, user.Balance); err != nil {
		return &InternalError{Message: "Error executing insert user db request", Err: err}
	}
	if err = stmt.Close(); err != nil {
		return &InternalError{Message: "Error when close db statement", Err: err}
	}
	return nil
}
