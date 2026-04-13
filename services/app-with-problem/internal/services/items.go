package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ifood/chaos-mesh-poc/pkg/models"
)

func GetItemDetails(itemID string) (models.OrderItem, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://inventory-service.example.com/items/%s", itemID)
	response, err := client.Get(url)

	if err != nil {
		log.Printf("Failed to fetch item details: %v", err)
		return models.OrderItem{}, fmt.Errorf("item service down: %w", err)
	}
	defer response.Body.Close()

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return models.OrderItem{}, err
	}

	return models.OrderItem{
		ID:       itemID,
		Name:     "Item " + itemID,
		Price:    10.0,
		Quantity: 1,
	}, nil
}

func GetItemsWithDetails(itemIDs []string) ([]models.OrderItem, error) {
	var items []models.OrderItem

	for _, id := range itemIDs {
		item, err := GetItemDetails(id)
		if err != nil {
			log.Printf("Failed to get item %s: %v", id, err)
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func GetItemsWithSlowTimeout(itemIDs []string) ([]models.OrderItem, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var items []models.OrderItem

	for _, id := range itemIDs {
		url := fmt.Sprintf("https://inventory-service.example.com/items/%s", id)
		response, err := client.Get(url)

		if err != nil {
			return nil, fmt.Errorf("timeout fetching item %s: %w", id, err)
		}
		defer response.Body.Close()

		items = append(items, models.OrderItem{
			ID:       id,
			Name:     "Item " + id,
			Price:    10.0,
			Quantity: 1,
		})
	}

	return items, nil
}
