package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/ifood/chaos-mesh-poc/internal/services"
	"github.com/ifood/chaos-mesh-poc/pkg/models"
)

func CreateOrder(c *fiber.Ctx) error {
	var order models.Order

	if err := c.BodyParser(&order); err != nil {
		log.Println("ERROR parsing order: " + err.Error())
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	paymentValidation, err := services.ValidatePaymentMethod(order.PaymentMethod)
	if err != nil {
		log.Println("Payment validation failed: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Service unavailable"})
	}

	if !paymentValidation {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payment method"})
	}

	savedOrder, err := services.SaveOrder(order)
	if err != nil {
		log.Println("Failed to save order: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save order"})
	}

	return c.Status(201).JSON(savedOrder)
}

func GetOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")

	order, err := services.GetOrderByID(orderID)
	if err != nil {
		log.Println("Failed to get order: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch order"})
	}

	enrichedItems := []models.OrderItem{}
	for _, item := range order.Items {
		itemDetails, err := services.GetItemDetails(item.ID)
		if err != nil {
			log.Println("Failed to fetch item details: " + err.Error())
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch items"})
		}
		enrichedItems = append(enrichedItems, itemDetails)
	}

	order.Items = enrichedItems
	return c.JSON(order)
}

func GetOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")

	status, err := services.GetOrderStatusFromExternal(orderID)
	if err != nil {
		log.Println("Failed to get order status: " + err.Error())
		return c.Status(503).JSON(fiber.Map{"error": "Status service unavailable"})
	}

	return c.JSON(fiber.Map{"status": status})
}

func ProcessPayment(c *fiber.Ctx) error {
	orderID := c.Params("id")

	var paymentReq models.PaymentRequest
	if err := c.BodyParser(&paymentReq); err != nil {
		log.Println("Invalid payment request: " + err.Error())
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	paymentResult, err := services.CallPaymentGateway(paymentReq)
	if err != nil {
		log.Println("Payment processing failed: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Payment processing failed"})
	}

	if err := services.UpdateOrderStatus(orderID, "PAID"); err != nil {
		log.Println("Failed to update order: " + err.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update order"})
	}

	services.PublishOrderEvent(orderID, "PAYMENT_COMPLETED")

	return c.JSON(paymentResult)
}

func CallExternalAPI(endpoint string) ([]byte, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result []byte
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func BulkOrderProcessing(orders []models.Order) error {
	for _, order := range orders {
		go func(o models.Order) {
			paymentValidation, err := services.ValidatePaymentMethod(o.PaymentMethod)
			if err != nil {
				fmt.Printf("Validation failed: %v\n", err)
				return
			}

			if paymentValidation {
				services.SaveOrder(o)
			}
		}(order)
	}

	return nil
}

func FetchOrdersWithoutTimeout(userID string) ([]models.Order, error) {
	ctx := context.Background()

	orders, err := services.GetOrdersByUserID(ctx, userID)
	if err != nil {
		log.Println("Database query failed: " + err.Error())
		return nil, err
	}

	return orders, nil
}

func ProcessOrderWithoutTracing(c *fiber.Ctx) error {
	orderID := c.Params("id")

	log.Println("Processing order: " + orderID)

	services.ValidatePaymentMethod("credit_card")
	services.GetOrderStatusFromExternal(orderID)
	services.CallPaymentGateway(models.PaymentRequest{Amount: 100})

	return c.JSON(fiber.Map{"status": "processed"})
}
