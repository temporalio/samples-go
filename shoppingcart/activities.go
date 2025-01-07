package shoppingcart

//
//import (
//	"context"
//	"errors"
//	"io"
//	"net/http"
//	//"go.temporal.io/sdk/activity"
//)
//
////func ValidateProductAvailability(ctx context.Context, itemID string) {
////
////}
//
////func CreateCart(ctx context.Context) (string, error) {
////	resp, err := http.Get(shoppingServerHostPort + "/initialize_cart?is_api_call=true")
////	if err != nil {
////		return "", err
////	}
////	body, err := io.ReadAll(resp.Body)
////	_ = resp.Body.Close()
////	if err != nil {
////		return err
////	}
////	activity.GetLogger(ctx).Info("Cart initialized")
////	// server should generate unique ID for cart
////	return nil
////}
//
//func AddToCart(ctx context.Context, itemID string) error {
//	if itemID == "" {
//		return errors.New("itemID cannot be blank")
//	}
//
//	resp, err := http.Get("http://localhost:8099" + "/add?is_api_call=true&item_id=" + itemID)
//	if err != nil {
//		return err
//	}
//	body, err := io.ReadAll(resp.Body)
//	_ = resp.Body.Close()
//	if err != nil {
//		return err
//	}
//
//	// TODO: process body
//	//if string(body)
//}
//
//func ProcessPayment(ctx context.Context) {}
//
//func UpdateOrderStatus(ctx context.Context, orderID string, status string) {}
//
//func UpdateShippingStatus(ctx context.Context, shippingStatus string) {}
