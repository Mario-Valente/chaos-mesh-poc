package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ifood/chaos-mesh-poc/internal/handlers"
)

func main() {
	app := fiber.New()

	api := app.Group("/api/v1")

	api.Post("/orders", handlers.CreateOrder)
	api.Get("/orders/:id", handlers.GetOrder)
	api.Get("/orders/:id/status", handlers.GetOrderStatus)
	api.Post("/orders/:id/pay", handlers.ProcessPayment)

	app.Listen(":3000")
}
