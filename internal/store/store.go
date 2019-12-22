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
	logger        *logrus.Logger
	db            *sql.DB
	users         map[uint64]*User
	userStatistic map[uint64]*Statistic
	// pendingActions PendingActions
}

type StoreHandler interface {
	GetUser(userID uint64) (*User, *Statistic, error)
	CreateUser(user *User) error
	CreateDeposit(d *Deposit) (float32, error)
}

type TransactionError struct {
	Err error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("%s", e.Err)
}

func (e *TransactionError) Unwrap() error {
	return e.Err
}

type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s", e.Err)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s", e.Err)
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

func New(ctx context.Context, logger *logrus.Logger, dbName string) StoreHandler {
	s := &Store{logger: logger}
	s.init(ctx, dbName)
	return s
}

func (s *Store) init(ctx context.Context, dbName string) {
	var err error
	s.db, err = sql.Open("sqlite3", dbName)
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

	s.initCache()

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

func (s *Store) initCache() {
	// init users list
	s.users = make(map[uint64]*User)
	userRows, err := s.db.Query("SELECT * FROM users")
	if err != nil {
		s.logger.Fatal("Can't load cached users: ", err)
	}
	defer userRows.Close()
	for userRows.Next() {
		var id uint64
		var balance float32
		err = userRows.Scan(&id, &balance)
		if err != nil {
			s.logger.Error("Can't read user from db: ", err)
		}
		s.users[id] = &User{
			ID:      id,
			Balance: balance,
		}
	}
	// init users statistic
	s.userStatistic = make(map[uint64]*Statistic)
	for _, user := range s.users {
		// TODO init statistic
		s.userStatistic[user.ID] = &Statistic{UserID: user.ID}
	}
}

// Create new user or return error
func (s *Store) CreateUser(user *User) error {
	// check, is user already exists
	if _, ok := s.users[user.ID]; ok {
		return &ValidationError{errors.New("User already exists")}
	}
	// check balance
	if user.Balance < 0 {
		return &ValidationError{errors.New("User balance may not be negative")}
	}
	stmt, err := s.db.Prepare("INSERT INTO users(id, balance) values(?, ?)")
	if err != nil {
		return &InternalError{Message: "Error when creating db statement", Err: err}
	}
	if _, err = stmt.Exec(user.ID, user.Balance); err != nil {
		return &InternalError{Message: "Error executing insert user db request", Err: err}
	}
	// add user to cache
	s.users[user.ID] = user
	s.userStatistic[user.ID] = &Statistic{UserID: user.ID}
	if err = stmt.Close(); err != nil {
		return &InternalError{Message: "Error when close db statement", Err: err}
	}
	return nil
}

func (s *Store) GetUser(userID uint64) (*User, *Statistic, error) {
	user, ok := s.users[userID]
	if !ok {
		return nil, nil, &NotFoundError{Err: errors.New("User not found")}
	}
	statistic, ok := s.userStatistic[userID]
	if !ok {
		s.logger.Warn("Not found statistic for user ", userID)
		return nil, nil, &NotFoundError{Err: errors.New("User not found")}
	}
	return user, statistic, nil
}

func (s *Store) CreateDeposit(d *Deposit) (float32, error) {
	if d.Amount == 0 {
		return 0, &ValidationError{errors.New("Deposit amount may be greater then zero")}
	}
	user, ok := s.users[d.UserID]
	if !ok {
		return 0, &NotFoundError{errors.New("User not found")}
	}
	stmt, err := s.db.Prepare("INSERT INTO deposits(id, user_id, balanceBefore, balanceAfter) values(?, ?, ?, ?)")
	if err != nil {
		return 0, &InternalError{Message: "Error when creating db statement", Err: err}
	}
	user.Lock()
	oldBalance := user.Balance
	newBalance := oldBalance + d.Amount
	if _, err = stmt.Exec(d.ID, d.UserID, oldBalance, newBalance); err != nil {
		user.Unlock()
		return 0, &TransactionError{Err: err}
	}
	user.Balance = newBalance
	statistic, ok := s.userStatistic[d.UserID]
	if ok {
		statistic.DepositeCount += 1
		statistic.DepositSum += d.Amount
	}
	user.Unlock()
	return newBalance, nil
}
