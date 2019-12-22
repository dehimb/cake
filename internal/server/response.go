package server

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserCreateResponse struct {
	Error string `json:"error"`
}

type UserResponse struct {
	UserID        uint64  `json:"id"`
	Balance       float32 `json:"balance"`
	DepositeCount int     `json:"depositCount"`
	DepositSum    float32 `json:"depositSum"`
	BetCount      int     `json:"betCount"`
	BetSum        float32 `json:"betSum"`
	WinCount      int     `json:"winCount"`
	WinSum        float32 `json:"winSum"`
}

type DepositResponse struct {
	Balance float32 `json:"balance"`
	Errror  string  `json:"error"`
}
