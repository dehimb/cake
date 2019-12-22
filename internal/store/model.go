package store

import "sync"

type User struct {
	sync.Mutex
	ID      uint64  `json:"id"`
	Balance float32 `json:"balance"`
}

type Statistic struct {
	sync.Mutex
	UserId        uint64
	DepositeCount int
	DepositSum    float32
	BetCount      int
	BetSum        float32
	WinCount      int
	WinSum        float32
}

type Deposit struct {
	ID     uint64
	UserID uint64
	Amount float32
}

type TransactionType string

const (
	Bet TransactionType = "Bet"
	Win TransactionType = "Win"
)

type Transaction struct {
	ID     uint64
	UserID uint64
	Type   TransactionType
	Amount float64
}

type PendingActions struct {
	sync.Mutex
	Deposits     []Deposit
	Transactions []Transaction
}
