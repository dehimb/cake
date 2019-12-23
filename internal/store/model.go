package store

import "sync"

type User struct {
	sync.Mutex
	ID      uint64  `json:"id"`
	Balance float32 `json:"balance"`
}

type Statistic struct {
	UserID        uint64
	DepositeCount int
	DepositSum    float32
	BetCount      int
	BetSum        float32
	WinCount      int
	WinSum        float32
}

type Deposit struct {
	ID     uint64  `json:"depositId"`
	UserID uint64  `json:"userId"`
	Amount float32 `json:"amount"`
}

type TransactionType string

const (
	Bet TransactionType = "Bet"
	Win TransactionType = "Win"
)

type Transaction struct {
	ID     uint64          `json:"transactionId"`
	UserID uint64          `json:"userId"`
	Type   TransactionType `json:"type"`
	Amount float32         `json:"amount"`
}

type PendingActions struct {
	sync.Mutex
	users []*User
}
