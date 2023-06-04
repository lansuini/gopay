package model

type LoginPassContent struct {
	AccountId int64  `json:"accountId"`
	Password  string `json:"password"`
}
