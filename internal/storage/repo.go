package storage

import "time"

type Repo interface {
	RegisterUser(credentials Credentials) error
	SelectUser(user Credentials) (Credentials, error)
	SignIn(user Credentials) (Credentials, error)
	PostOrder(order Order) error
	GetOrderByOrderID(orderid int64) (Order, error)
	GetOrdersByUserID(userid int64) ([]Order, error)
	GetBalanceByUserID(userid int64) (float64, error)
	PostWithdrawal(withdrawal Withdrawal) error
	SetBalanceByUserID(userid int64, current float64) error
}

type Credentials struct {
	ID       int
	Username string `json:"username"`
	Password string `json:"password"`
}

type Order struct {
	OrderID    int64     `json:"order"`
	UserID     int       `json:"userID,omitempty"`
	State      string    `json:"state,omitempty"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at,omitempty"`
}

type Balance struct {
	UserID  int     `json:"-"`
	Current float64 `json:"current"`
	//Withdrawn float32 `json:"withdrawn"`
}
type Withdrawal struct {
	UserID      int       `json:"-"`
	OrderID     int64     `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
