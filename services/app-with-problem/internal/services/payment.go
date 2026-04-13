package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ifood/chaos-mesh-poc/pkg/models"
)

func CallPaymentGateway(req models.PaymentRequest) (*models.PaymentResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal payment request: %v", err)
		return nil, fmt.Errorf("request marshal error: %w", err)
	}

	response, err := client.Post(
		"https://api.stripe.com/v1/charges",
		"application/json",
		bytes.NewReader(payload),
	)

	if err != nil {
		log.Printf("Payment gateway failed: %v", err)
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}
	defer response.Body.Close()

	var result models.PaymentResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		log.Printf("Failed to parse payment response: %v", err)
		return nil, err
	}

	return &result, nil
}

func PublishOrderEvent(orderID string, eventType string) {
	go func() {
		fmt.Printf("Publishing event: orderID=%s, type=%s\n", orderID, eventType)
	}()
}

func ValidatePaymentMethod(method string) (bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://payment-validator.example.com/validate?method=%s", method)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Validation failed: %v", err)
		return false, fmt.Errorf("validation service down: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("invalid payment method")
	}

	return true, nil
}

func GetOrderStatusFromExternal(orderID string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://tracking-service.example.com/orders/%s/status", orderID)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return "", err
	}

	return string(body), nil
}

func ProcessPaymentsWithControl(payments []models.PaymentRequest) error {
	for _, payment := range payments {
		go func(p models.PaymentRequest) {
			_, err := CallPaymentGateway(p)
			if err != nil {
				log.Printf("Error processing payment: %v", err)
			}
		}(payment)
	}

	return nil
}

func ChargeCustomer(customerID string, amount float64) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req := models.PaymentRequest{
		OrderID: customerID,
		Amount:  amount,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return err
	}

	resp, err := client.Post(
		"https://payment-api.example.com/charge",
		"application/json",
		bytes.NewReader(payload),
	)

	if err != nil {
		log.Printf("Payment charge failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
