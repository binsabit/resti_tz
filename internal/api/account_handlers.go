package api

import (
	"errors"
	"github.com/binsabit/resti_tz/internal/account"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

var MsgAccountExists = "Account already exists"
var MsgNegativeBalance = "Negative balance"

func (h Handler) RegisterAccountRoutes(app *fiber.App) {

	app.Group("/api/v1/account").
		Post("/", h.createAccountHandler).
		Get("/:accountId", h.getAccountHandler)

}

func (h Handler) createAccountHandler(c *fiber.Ctx) error {

	var input struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}

	if err := c.BodyParser(&input); err != nil {
		log.Println("error parsing body", input, err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	ctx := c.Context()

	exist, err := h.accountRepo.CheckIfAccountNameExists(ctx, h.db.GetDb(), input.Name)
	if err != nil {
		log.Println("error while checking if account exists", input, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if exist {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": MsgAccountExists})
	}

	if input.Balance < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": MsgNegativeBalance})
	}

	acc, err := h.accountRepo.SaveAccount(ctx, h.db.GetDb(), account.Account{
		Name:    &input.Name,
		Balance: &input.Balance,
	})
	if err != nil {
		log.Println("error while saving account", input, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(acc)

}

func (h Handler) getAccountHandler(c *fiber.Ctx) error {
	accountId, err := strconv.ParseInt(c.Params("accountId"), 10, 64)
	if err != nil {
		log.Println("error parsing id", c, err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	ctx := c.Context()

	acc, err := h.accountRepo.GetAccount(ctx, h.db.GetDb(), accountId)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		log.Println("error getting account", accountId, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	transactions, err := h.transactionRepo.GetAllTransactionsWithAccountID(ctx, h.db.GetDb(), accountId)
	if err != nil {
		log.Println("error getting transactions", accountId, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"account": acc, "transactions": transactions})
}
