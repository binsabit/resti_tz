package api

import (
	"context"
	"errors"
	"github.com/binsabit/resti_tz/internal/account"
	"github.com/binsabit/resti_tz/internal/transactions"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func (h Handler) RegisterTransactionRoutes(app *fiber.App) {

	app.Group("/api/v1/transaction").
		Post("", h.createTransactionHandler).
		Get("all", h.getAllTransactionsHandler).
		Get("/account/:accountId", h.getAllTransactionsForAccountHandler)
	//Get("/:transactionId")

}

func (h Handler) createTransactionHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	var input struct {
		Account1  int64                  `json:"account1"`
		Account2  *int64                 `json:"account2"`
		Amount    float64                `json:"amount"`
		Operation transactions.Operation `json:"operation"`
	}
	if err := c.BodyParser(&input); err != nil {
		log.Println("error parsing body", input, err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	tr := transactions.Transaction{
		Account1: account.Account{
			Id: input.Account1,
		},
		Amount:    input.Amount,
		Operation: input.Operation,
	}

	if input.Account2 != nil {
		tr.Account2 = &account.Account{Id: *input.Account2}
	}

	err := func(ctx context.Context) error {
		tx, err := h.db.BeginTx(ctx)
		if err != nil {
			return err
		}

		defer tx.Rollback(ctx)

		tr, err = h.transactionRepo.CreateTransaction(ctx, tx, tr)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		return nil
	}(ctx)

	if err != nil {
		switch {
		case errors.Is(err, transactions.ErrAccountIdMissing) ||
			errors.Is(err, transactions.ErrAccountIdMissing) ||
			errors.Is(err, transactions.ErrNotEnoughBalance):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		log.Println("error creating transaction", input, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"transaction": tr})
}

func (h Handler) getAllTransactionsHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	transactionsList, err := h.transactionRepo.GetAllTransactions(ctx, h.db.GetDb())
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"transactions": transactionsList})
}

func (h Handler) getAllTransactionsForAccountHandler(c *fiber.Ctx) error {
	accountId, err := strconv.ParseInt(c.Params("accountId"), 10, 64)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	ctx := c.Context()

	transactionsList, err := h.transactionRepo.GetAllTransactionsWithAccountID(ctx, h.db.GetDb(), accountId)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"transactions": transactionsList})
}
