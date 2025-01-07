package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.temporal.io/sdk/client"
	"net/http"
	"os"
	"sort"
	"sync"
)

var (
	cartState      = make(map[string]int) // id -> itemName -> number
	workflowClient client.Client
	itemCosts      = map[string]int{
		"apple":      2,
		"banana":     1,
		"watermelon": 5,
		"television": 1000,
		"house":      10000000,
		"car":        50000,
		"binder":     10,
	}
)

type WebSocketServer struct {
	clients   map[string]*websocket.Conn
	mu        sync.Mutex
	broadcast chan WebSocketMessage
}

type WebSocketMessage struct {
	UserID string `json:"user_id"`
	Event  string `json:"event"`
	Data   any    `json:"data"`
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[string]*websocket.Conn),
		broadcast: make(chan WebSocketMessage),
	}
}

func (s *WebSocketServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Assume a user ID is passed as a query param for simplicity
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		fmt.Println("Missing user_id in query parameters")
		return
	}

	// Register the client
	s.mu.Lock()
	s.clients[userID] = conn
	s.mu.Unlock()

	fmt.Printf("Client connected: %s\n", userID)

	// Keep connection open until closed by the client
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}

	// Unregister the client
	s.mu.Lock()
	delete(s.clients, userID)
	s.mu.Unlock()
	fmt.Printf("Client disconnected: %s\n", userID)
}

func (s *WebSocketServer) handleMessages() {
	for msg := range s.broadcast {
		s.mu.Lock()
		conn, exists := s.clients[msg.UserID]
		s.mu.Unlock()
		if !exists {
			fmt.Printf("User %s not connected\n", msg.UserID)
			continue
		}

		if err := conn.WriteJSON(msg); err != nil {
			fmt.Printf("Error sending message to user %s: %v\n", msg.UserID, err)
			s.mu.Lock()
			delete(s.clients, msg.UserID)
			s.mu.Unlock()
		}
	}
}

func (s *WebSocketServer) SendMessage(userID, event string, data any) {
	s.broadcast <- WebSocketMessage{UserID: userID, Event: event, Data: data}
}

func main() {
	wsServer := NewWebSocketServer()

	var err error
	workflowClient, err = client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting dummy server...")
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/ws", wsServer.handleConnections)

	go wsServer.handleMessages()

	fmt.Println("WebSocket server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting WebSocket server:", err)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	// read in javascript that handles websocket
	fileContents, err := os.ReadFile("shoppingcart/home.html")
	if err != nil {
		http.Error(w, "Could not read shoppingcart/home.html", http.StatusInternalServerError)
		fmt.Println("Error reading shoppingcart/home.html:", err)
		return
	}

	// Write the contents to the HTTP response
	w.Header().Set("Content-Type", "text/html") // Set the content type to HTML
	_, _ = fmt.Fprint(w, "<!DOCTYPE html><html>")
	_, _ = fmt.Fprintf(w, "%s", fileContents)
	_, _ = fmt.Fprint(w, "<h1>DUMMY SHOPPING WEBSITE</h1>"+
		"<a href=\"/list\">HOME</a>"+
		"<a href=\"/list\">TODO:Payment</a>"+
		"<a href=\"/list\">TODO:Shipment</a>"+
		"<h3>Available Items to Purchase</h3><table border=1><tr><th>Item</th><th>Cost</th><th>Action</th>")

	//	and at the end of the workflow the server will send return data to the client/website
	keys := make([]string, 0)
	count := 0
	for k, _ := range itemCosts {
		count += 1
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		//actionButton := fmt.Sprintf("<a href=\"/action?type=add&id=%s\">"+
		//	"<button style=\"background-color:#4CAF50;\">Add to Cart</button></a>", k)
		actionButton := fmt.Sprintf("<button "+
			"style=\"background-color:#4CAF50;\" "+
			"onclick=\"sendAddToCartRequest('%s')\">"+
			"Add to Cart</button></a>", k)

		_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td><td>%s</td></tr>", k, itemCosts[k], actionButton)
	}

	_, _ = fmt.Fprint(w, "</table><div id=\"cart\"></div></html>")

	//_, _ = fmt.Fprint(w, "</table><h3>Current items in cart:</h3>"+
	//	"<table border=1><tr><th>Item</th><th>Quantity</th><th>Action</th>")
	//
	//// TODO: List current items in cart
	//// TODO: query from websocket?
	//for key, val := range cartState {
	//	// TODO: add remove action
	//	_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td><td></td></tr>", key, val)
	//}
	//_, _ = fmt.Fprint(w, "</table>")
}
