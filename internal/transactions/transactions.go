package transactions

import (
	"context"
	"database/sql"
	"errors"
	"github.com/binsabit/resti_tz/internal/account"
	database "github.com/binsabit/resti_tz/internal/repository"
	"github.com/jackc/pgx/v5"
	"time"
)

var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrAccountIdMissing = errors.New("account id is missing")

type Operation string

const WithdrawalOp Operation = "withdrawal"
const DepositOp Operation = "deposit"
const TransferOp Operation = "transfer"

type Transaction struct {
	Id        int64            `json:"id"`
	Account1  account.Account  `json:"account1"`
	Account2  *account.Account `json:"account2,omitempty"`
	Amount    float64          `json:"amount"`
	Operation Operation        `json:"operation"`
	CreatedAt time.Time        `json:"createdAt"`
}

type Repository struct {
	account.Repository
}

func (r Repository) CreateTransaction(ctx context.Context, db database.DBTX, transaction Transaction) (Transaction, error) {

	err := r.AccountMustExist(ctx, db, transaction.Account1.Id)
	if err != nil {
		return Transaction{}, err
	}

	switch transaction.Operation {
	case WithdrawalOp:
		return r.Withdrawal(ctx, db, transaction.Account1.Id, transaction.Amount)
	case DepositOp:
		return r.Deposit(ctx, db, transaction.Account1.Id, transaction.Amount)
	case TransferOp:
		if transaction.Account2 == nil {
			return Transaction{}, ErrAccountIdMissing
		}
		err := r.AccountMustExist(ctx, db, transaction.Account2.Id)
		if err != nil {
			return Transaction{}, err
		}
		return r.Transfer(ctx, db, transaction.Account1.Id, transaction.Account2.Id, transaction.Amount)
	}
	return Transaction{}, nil
}

func (r Repository) Withdrawal(ctx context.Context, db database.DBTX, account1 int64, amount float64) (Transaction, error) {

	balance, err := r.GetAccountBalance(ctx, db, account1)
	if err != nil {
		return Transaction{}, err
	}

	if balance < amount {
		return Transaction{}, ErrNotEnoughBalance
	}

	balance = balance - amount

	err = r.UpdateBalance(ctx, db, account1, balance)
	if err != nil {
		return Transaction{}, err
	}

	return r.SaveTransaction(ctx, db, createTransactionParams{
		account1:  account1,
		amount:    amount,
		operation: WithdrawalOp,
	})

}

func (r Repository) Deposit(ctx context.Context, db database.DBTX, account1 int64, amount float64) (Transaction, error) {
	balance, err := r.GetAccountBalance(ctx, db, account1)
	if err != nil {
		return Transaction{}, err
	}

	balance = balance + amount

	err = r.UpdateBalance(ctx, db, account1, balance)
	if err != nil {
		return Transaction{}, err
	}

	return r.SaveTransaction(ctx, db, createTransactionParams{
		account1:  account1,
		amount:    amount,
		operation: DepositOp,
	})
}

func (r Repository) Transfer(ctx context.Context, db database.DBTX, account1, account2 int64, amount float64) (Transaction, error) {
	balance1, err := r.GetAccountBalanceForUpdate(ctx, db, account1)
	if err != nil {
		return Transaction{}, err
	}
	if balance1 < amount {
		return Transaction{}, ErrNotEnoughBalance
	}

	balance2, err := r.GetAccountBalanceForUpdate(ctx, db, account2)
	if err != nil {
		return Transaction{}, err
	}

	balance2 = balance2 + amount

	balance1 = balance1 - amount

	err = r.UpdateBalance(ctx, db, account1, balance1)
	if err != nil {
		return Transaction{}, err
	}

	err = r.UpdateBalance(ctx, db, account2, balance2)
	if err != nil {
		return Transaction{}, err
	}

	return r.SaveTransaction(ctx, db, createTransactionParams{
		account1:  account1,
		account2:  &account2,
		amount:    amount,
		operation: TransferOp,
	})
}

type createTransactionParams struct {
	account1  int64
	account2  *int64
	amount    float64
	operation Operation
}

func (r Repository) SaveTransaction(ctx context.Context, db database.DBTX, params createTransactionParams) (Transaction, error) {
	query := `insert into transactions(account1,account2,amount,operation) values($1,$2,$3,$4) returning id,created_at`

	var (
		id        int64
		createdAt time.Time
	)

	args := []interface{}{params.account1, params.account2, params.amount, params.operation}

	err := db.QueryRow(ctx, query, args...).Scan(&id, &createdAt)

	res := Transaction{
		Id:        id,
		Account1:  account.Account{Id: params.account1},
		Amount:    params.amount,
		Operation: params.operation,
		CreatedAt: createdAt,
	}
	if params.operation == TransferOp {
		res.Account2 = &account.Account{Id: *params.account2}
	}
	return res, err
}

func (r Repository) GetAllTransactions(ctx context.Context, db database.DBTX) ([]Transaction, error) {
	var transactions []Transaction

	query := `SELECT t.id AS transaction_id, t.amount, t.operation,t.created_at, a1.name AS account_name,a1.id,
			CASE  
			    WHEN t.operation = 'transfer' THEN a2.name ELSE NULL
			END AS account2_name,
			CASE  
			    WHEN t.operation = 'transfer' THEN a2.id ELSE NULL
			END AS account2_id
			FROM  transactions t
			LEFT JOIN accounts a1 ON t.account1 = a1.id
			LEFT JOIN accounts a2 ON t.account2 = a2.id
			ORDER BY 
			t.created_at DESC;`

	rows, err := db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			t           Transaction
			accountName sql.NullString
			accountId   sql.NullInt64
		)
		err := rows.Scan(&t.Id, &t.Amount, &t.Operation, &t.CreatedAt, &t.Account1.Name, &t.Account1.Id, &accountName, &accountId)
		if err != nil {
			return nil, err
		}

		if accountId.Valid {
			t.Account1.Id = accountId.Int64
		}
		if accountName.Valid {
			t.Account2.Name = &accountName.String
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (r Repository) GetAllTransactionsWithAccountID(ctx context.Context, db database.DBTX, accountId int64) ([]Transaction, error) {
	var transactions []Transaction

	query := `SELECT t.id AS transaction_id, t.amount, t.operation,t.created_at, a1.name AS account_name,a1.id,
			CASE WHEN t.operation = 'transfer' THEN a2.name ELSE NULL END AS account2_name,
			CASE WHEN t.operation = 'transfer' THEN a2.id ELSE NULL END AS account2_id
			FROM  transactions t
			LEFT JOIN accounts a1 ON t.account1 = a1.id
			LEFT JOIN accounts a2 ON t.account2 = a2.id
			    where a1.id = $1 or a2.id = $1 
			ORDER BY 
			t.created_at DESC;`

	rows, err := db.Query(ctx, query, accountId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			t           Transaction
			accountName sql.NullString
			accountId   sql.NullInt64
		)
		err := rows.Scan(&t.Id, &t.Amount, &t.Operation, &t.CreatedAt, &t.Account1.Name, &t.Account1.Id, &accountName, &accountId)
		if err != nil {
			return nil, err
		}

		if accountId.Valid {
			t.Account2 = new(account.Account)
			t.Account2.Id = accountId.Int64
		}
		if accountName.Valid {

			t.Account2.Name = &accountName.String
		}

		transactions = append(transactions, t)
	}
	return transactions, nil
}
