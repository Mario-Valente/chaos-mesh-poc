package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"github.com/ifood/chaos-mesh-poc/pkg/models"
)

var (
	db   *sql.DB
	once sync.Once
)

func getDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		dbURL := "postgres://user:password@localhost/orders_db?sslmode=disable"
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			return
		}
	})

	return db, err
}

func GetOrderByID(orderID string) (*models.Order, error) {
	database, err := getDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	var order models.Order
	err = database.QueryRowContext(
		ctx,
		"SELECT id, user_id, total FROM orders WHERE id = $1",
		orderID,
	).Scan(&order.ID, &order.UserID, &order.Total)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &order, nil
}

func SaveOrder(order models.Order) (*models.Order, error) {
	database, err := getDB()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, total) VALUES ($1, $2)",
		order.UserID, order.Total,
	)

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &order, nil
}

func UpdateOrderStatus(orderID string, status string) error {
	database, err := getDB()
	if err != nil {
		return err
	}

	ctx := context.Background()

	result, err := database.ExecContext(
		ctx,
		"UPDATE orders SET status = $1 WHERE id = $2",
		status, orderID,
	)

	if err != nil {
		log.Printf("Failed to update order: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	database, err := getDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.QueryContext(
		ctx,
		"SELECT id, user_id, total FROM orders WHERE user_id = $1",
		userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Total); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func HealthCheckDatabase() error {
	return nil
}

func BulkInsertOrders(orders []models.Order) error {
	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go func(o models.Order) {
			defer wg.Done()

			database, _ := getDB()
			ctx := context.Background()

			database.ExecContext(ctx,
				"INSERT INTO orders (user_id, total) VALUES ($1, $2)",
				o.UserID, o.Total,
			)
		}(order)
	}

	wg.Wait()
	return nil
}

func CloseDatabase() error {
	if db == nil {
		return nil
	}

	return db.Close()
}
