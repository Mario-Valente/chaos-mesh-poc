package models

type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	Items         []OrderItem `json:"items"`
	Total         float64     `json:"total"`
	Status        string      `json:"status"`
	PaymentMethod string      `json:"payment_method"`
}

type OrderItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
}

type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Amount        float64 `json:"amount"`
}
