package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	shoppingcart "github.com/temporalio/samples-go/shoppingcart"
	"go.temporal.io/sdk/client"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust for production
	},
}

// WebSocketMessage defines the structure of the message sent by the web app
type WebSocketMessage struct {
	Action   string `json:"action"` // "add" or "remove"
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type CartStatusMessage struct {
	Action string    `json:"action"`
	Data   CartState `json:"data"`
}

// CartSignalPayload is the payload structure for Temporal signals
//type CartSignalPayload struct {
//	Action   string `json:"action"` // "add" or "remove"
//	ItemID   string `json:"item_id"`
//	Quantity int    `json:"quantity"`
//}

// WebSocketServer holds the WebSocket connections and Temporal client
type WebSocketServer struct {
	connections    map[string]*websocket.Conn // Map of user_id to WebSocket connection
	mu             sync.Mutex
	temporalClient client.Client
}

type CartState map[string]int // itemID -> quantity

// NewWebSocketServer creates a new WebSocket server instance
func NewWebSocketServer(temporalClient client.Client) *WebSocketServer {
	return &WebSocketServer{
		connections:    make(map[string]*websocket.Conn),
		temporalClient: temporalClient,
	}
}

// HandleConnections manages incoming WebSocket connections
func (s *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.Println("user_id is missing in the query parameters")
		return
	}

	// Register the connection
	s.mu.Lock()
	s.connections[userID] = conn
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.connections, userID)
		s.mu.Unlock()
	}()

	log.Printf("WebSocket connection established for user: %s", userID)

	// Handle incoming messages
	for {
		log.Println("calling conn.ReadMessage()")
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}

		// Handle the WebSocket message
		s.handleMessage(userID, message)
	}
}

// handleMessage processes incoming WebSocket messages and triggers Temporal signals
func (s *WebSocketServer) handleMessage(userID string, message []byte) {
	// Parse the WebSocket message
	var msg WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error parsing WebSocket message:", err)
		return
	}
	workflowID := fmt.Sprintf("shopping_cart_%s", userID)

	switch msg.Action {
	case "add":
		// Signal to add an item to the cart
		signalPayload := shoppingcart.CartSignalPayload{
			Action:   "add",
			ItemID:   msg.ItemID,
			Quantity: msg.Quantity,
		}
		log.Println("Sending signal payload", workflowID, signalPayload)
		err := s.temporalClient.SignalWorkflow(context.Background(), workflowID, "", "cart_signal", signalPayload)
		if err != nil {
			log.Println("Error signaling workflow:", err)
		}

		// TODO: query the cart and push signal back to webapp
		var cartState CartState
		resp, err := s.temporalClient.QueryWorkflow(context.Background(), workflowID, "", "get_cart")
		if err != nil {
			log.Println("Error querying workflow:", err)
			return
		}
		if err := resp.Get(&cartState); err != nil {
			log.Fatalln("Unable to decode query result", err)
		}

		// Send the cart state back to the WebSocket client
		response := CartStatusMessage{
			Action: "cart_state",
			Data:   cartState,
		}
		conn := s.connections[userID]
		if conn != nil {
			conn.WriteJSON(response)
		}
	case "get_cart":
		// Query the cart state
		var cartState CartState
		resp, err := s.temporalClient.QueryWorkflow(context.Background(), workflowID, "", "get_cart")
		if err != nil {
			log.Println("Error querying workflow:", err)
			return
		}
		if err := resp.Get(&cartState); err != nil {
			log.Fatalln("Unable to decode query result", err)
		}

		// Send the cart state back to the WebSocket client
		response := CartStatusMessage{
			Action: "cart_state",
			Data:   cartState,
		}
		conn := s.connections[userID]
		if conn != nil {
			conn.WriteJSON(response)
		}
	default:
		log.Printf("Unknown action: %s\n", msg.Action)
	}
}

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Error creating Temporal client: %v", err)
	}
	defer c.Close()

	server := NewWebSocketServer(c)

	http.HandleFunc("/ws", server.HandleConnections)
	log.Println("WebSocket server is running on ws://localhost:8089/ws")
	if err := http.ListenAndServe(":8089", nil); err != nil {
		log.Fatalf("Error starting WebSocket server: %v", err)
	}
}
