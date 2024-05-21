package http

import (
	"github.com/binsabit/resti_tz/config"
	"github.com/binsabit/resti_tz/internal/api"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	router *fiber.App
	port   string
}

func NewServer(cfg config.Http, handler *api.Handler) *Server {
	router := fiber.New()

	handler.RegisterTransactionRoutes(router)

	handler.RegisterAccountRoutes(router)

	return &Server{
		router: router,
		port:   cfg.Port,
	}
}

func (s *Server) Start() error {

	return s.router.Listen(":" + s.port)
}

func (s *Server) Stop() error {
	return s.router.Shutdown()
}
