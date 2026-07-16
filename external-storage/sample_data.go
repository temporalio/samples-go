package externalstorage

import (
	"fmt"
	"hash/fnv"
	"math/rand"
)

// Produce payloads large enough to exceed the default 256 KiB ExternalStorage
// threshold without hand-crafted catalogs. Each order is padded with random
// filler in its item descriptions and shipping notes. Calibrated so 100 orders
// serialize to roughly 300 KiB of JSON.

const (
	itemsPerOrder         = 5
	itemDescriptionChars  = 500
	shippingNotesChars    = 200
)

var cities = []struct {
	city, state string
}{
	{"Houston", "TX"},
	{"Dallas", "TX"},
	{"Los Angeles", "CA"},
	{"San Francisco", "CA"},
	{"Denver", "CO"},
	{"Miami", "FL"},
	{"Chicago", "IL"},
	{"New York", "NY"},
	{"Seattle", "WA"},
	{"Atlanta", "GA"},
}

const fillerAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ "

func filler(rng *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = fillerAlphabet[rng.Intn(len(fillerAlphabet))]
	}
	return string(b)
}

func generateOrders(batchID string, count int) []Order {
	orders := make([]Order, count)
	for i := 0; i < count; i++ {
		orders[i] = generateOrder(batchID, i+1)
	}
	return orders
}

func generateOrder(batchID string, index int) Order {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s-%d", batchID, index)
	rng := rand.New(rand.NewSource(int64(h.Sum64())))
	loc := cities[rng.Intn(len(cities))]

	items := make([]OrderItem, itemsPerOrder)
	var totalWeight float64
	for i := range items {
		qty := rng.Intn(10) + 1
		weight := round2(0.5 + rng.Float64()*49.5)
		items[i] = OrderItem{
			SKU:          fmt.Sprintf("SKU-%05d", 10000+rng.Intn(90000)),
			Name:         fmt.Sprintf("Product %d", 1+rng.Intn(999)),
			Description:  filler(rng, itemDescriptionChars),
			Quantity:     qty,
			UnitPriceUSD: round2(10.0 + rng.Float64()*990.0),
			WeightKg:     weight,
		}
		totalWeight += weight * float64(qty)
	}

	return Order{
		ID: fmt.Sprintf("ORD-%06d", index),
		Customer: Customer{
			ID:    fmt.Sprintf("CUST-%06d", index),
			Name:  fmt.Sprintf("Customer %d", index),
			Email: fmt.Sprintf("customer%d@example.com", index),
			Address: Address{
				Street:  fmt.Sprintf("%d Main Street", 100+rng.Intn(9900)),
				City:    loc.city,
				State:   loc.state,
				ZipCode: fmt.Sprintf("%05d", 10000+rng.Intn(90000)),
				Country: "US",
			},
		},
		Items:         items,
		TotalWeightKg: round2(totalWeight),
		ShippingNotes: filler(rng, shippingNotesChars),
	}
}

