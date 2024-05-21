package account

import (
	"context"
	"errors"
	database "github.com/binsabit/resti_tz/internal/repository"
	"github.com/jackc/pgx/v5"
)

var ErrAccountNotFound = errors.New("account not found")

type Account struct {
	Id      int64    `json:"id"`
	Name    *string  `json:"name,omitempty"`
	Balance *float64 `json:"balance,omitempty"`
}

type Repository struct {
}

func (a Repository) AccountMustExist(ctx context.Context, db database.DBTX, accountId int64) error {
	query := `select count(*) from accounts where id=$1`
	var count int
	err := db.QueryRow(ctx, query, accountId).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrAccountNotFound
	}
	return nil
}

func (a Repository) CheckIfAccountNameExists(ctx context.Context, db database.DBTX, name string) (bool, error) {
	query := `select count(*) from accounts where name=$1`
	var count int
	err := db.QueryRow(ctx, query, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (a Repository) SaveAccount(ctx context.Context, db database.DBTX, account Account) (Account, error) {
	query := `insert into accounts (name, balance) values ($1, $2) returning id`

	var id int64

	err := db.QueryRow(ctx, query, account.Name, account.Balance).Scan(&id)
	account.Id = id
	return account, err
}

func (a Repository) GetAccountBalance(ctx context.Context, db database.DBTX, accountId int64) (float64, error) {
	query := `select balance from accounts where id=$1`
	var balance float64
	err := db.QueryRow(ctx, query, accountId).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, err
	}
	return balance, nil

}
func (a Repository) GetAccountBalanceForUpdate(ctx context.Context, db database.DBTX, accountId int64) (float64, error) {
	query := `select balance from accounts where id=$1 for update`
	var balance float64
	err := db.QueryRow(ctx, query, accountId).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, err
	}
	return balance, nil
}

func (a Repository) UpdateBalance(ctx context.Context, db database.DBTX, accountId int64, balance float64) error {

	query := `update accounts set balance = $1 where id=$2`

	_, err := db.Exec(ctx, query, balance, accountId)

	return err
}

func (a Repository) GetAccount(ctx context.Context, db database.DBTX, accountId int64) (Account, error) {
	var account Account

	query := `select id,name,balance from accounts where id=$1`

	err := db.QueryRow(ctx, query, accountId).Scan(&account.Id, &account.Name, &account.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, ErrAccountNotFound
		}
		return account, err
	}
	return account, nil
}
