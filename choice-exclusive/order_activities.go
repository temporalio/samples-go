package choice

import (
	"context"
	"fmt"
	"math/rand"

	"go.temporal.io/sdk/activity"
)

type OrderActivities struct {
	OrderChoices []string
}

func (a *OrderActivities) GetOrder() (string, error) {
	idx := rand.Intn(len(a.OrderChoices))
	order := a.OrderChoices[idx]
	fmt.Printf("Order is for %s\n", order)
	return order, nil
}

func (a *OrderActivities) OrderApple(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderBanana(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderCherry(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) OrderOrange(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func (a *OrderActivities) GetBasketOrder(ctx context.Context) ([]string, error) {
	var basket []string
	for _, item := range a.OrderChoices {
		// some random decision
		if rand.Float32() <= 0.65 {
			basket = append(basket, item)
		}
	}

	if len(basket) == 0 {
		basket = append(basket, a.OrderChoices[rand.Intn(len(a.OrderChoices))])
	}

	activity.GetLogger(ctx).Info("Get basket order.", "Orders", basket)
	return basket, nil
}
