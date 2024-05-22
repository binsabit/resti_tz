package api

import (
	"github.com/binsabit/resti_tz/internal/account"
	database "github.com/binsabit/resti_tz/internal/repository"
	"github.com/binsabit/resti_tz/internal/transactions"
)

type Handler struct {
	db              *database.Database
	transactionRepo transactions.Repository
	accountRepo     account.Repository
}

func NewHandler(db *database.Database, transactionRepo transactions.Repository, accountRepo account.Repository) *Handler {
	return &Handler{
		db:              db,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}
