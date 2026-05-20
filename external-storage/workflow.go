package externalstorage

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	// TaskQueue is the task queue used by the external-storage worker and starter.
	TaskQueue = "external-storage-task-queue"

	// warehouseState is the fulfillment center state used to estimate delivery.
	warehouseState = "TX"
)

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type Customer struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
}

type OrderItem struct {
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Quantity     int     `json:"quantity"`
	UnitPriceUSD float64 `json:"unit_price_usd"`
	WeightKg     float64 `json:"weight_kg"`
}

type Order struct {
	ID            string      `json:"id"`
	Customer      Customer    `json:"customer"`
	Items         []OrderItem `json:"items"`
	TotalWeightKg float64     `json:"total_weight_kg"`
	ShippingNotes string      `json:"shipping_notes"`
}

type ProcessedOrder struct {
	ID                    string  `json:"id"`
	CustomerID            string  `json:"customer_id"`
	DestinationCity       string  `json:"destination_city"`
	DestinationState      string  `json:"destination_state"`
	TotalWeightKg         float64 `json:"total_weight_kg"`
	ShippingCostUSD       float64 `json:"shipping_cost_usd"`
	EstimatedDeliveryDays int     `json:"estimated_delivery_days"`
}

type OrderBatchRequest struct {
	BatchID    string `json:"batch_id"`
	OrderCount int    `json:"order_count"`
}

type BatchSummary struct {
	BatchID              string  `json:"batch_id"`
	OrderCount           int     `json:"order_count"`
	TotalShippingCostUSD float64 `json:"total_shipping_cost_usd"`
	TotalWeightKg        float64 `json:"total_weight_kg"`
	AvgDeliveryDays      float64 `json:"avg_delivery_days"`
}

// FetchOrders returns the orders for a batch. The slice is intentionally large
// enough that, even after zlib compression, it exceeds the default
// ExternalStorage threshold and gets offloaded to S3.
func FetchOrders(_ context.Context, request OrderBatchRequest) ([]Order, error) {
	return generateOrders(request.BatchID, request.OrderCount), nil
}

// ProcessOrders computes a shipping cost and an estimated delivery day count
// for each order, and returns the per-order results.
func ProcessOrders(_ context.Context, orders []Order) ([]ProcessedOrder, error) {
	results := make([]ProcessedOrder, len(orders))
	for i, order := range orders {
		days := 5
		if order.Customer.Address.State == warehouseState {
			days = 2
		}
		results[i] = ProcessedOrder{
			ID:                    order.ID,
			CustomerID:            order.Customer.ID,
			DestinationCity:       order.Customer.Address.City,
			DestinationState:      order.Customer.Address.State,
			TotalWeightKg:         order.TotalWeightKg,
			ShippingCostUSD:       round2(2.50 + 1.20*order.TotalWeightKg),
			EstimatedDeliveryDays: days,
		}
	}
	return results, nil
}

// ProcessOrderBatchWorkflow fetches a batch of orders and processes them.
// The workflow input and result are small. The intermediate order list (the
// output of FetchOrders and the input of ProcessOrders) is the payload that
// exceeds the size threshold and gets offloaded to external storage.
func ProcessOrderBatchWorkflow(ctx workflow.Context, request OrderBatchRequest) (BatchSummary, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	})

	var orders []Order
	if err := workflow.ExecuteActivity(ctx, FetchOrders, request).Get(ctx, &orders); err != nil {
		return BatchSummary{}, err
	}

	var processed []ProcessedOrder
	if err := workflow.ExecuteActivity(ctx, ProcessOrders, orders).Get(ctx, &processed); err != nil {
		return BatchSummary{}, err
	}

	var totalCost, totalWeight float64
	var totalDays int
	for _, p := range processed {
		totalCost += p.ShippingCostUSD
		totalWeight += p.TotalWeightKg
		totalDays += p.EstimatedDeliveryDays
	}
	var avgDays float64
	if len(processed) > 0 {
		avgDays = float64(totalDays) / float64(len(processed))
	}

	return BatchSummary{
		BatchID:              request.BatchID,
		OrderCount:           len(processed),
		TotalShippingCostUSD: round2(totalCost),
		TotalWeightKg:        round2(totalWeight),
		AvgDeliveryDays:      round1(avgDays),
	}, nil
}

func round1(v float64) float64 { return float64(int(v*10+0.5)) / 10 }
func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
