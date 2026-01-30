// WebSocket Test Client
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

func main() {
	fmt.Println("=== WebSocket Test Client ===")

	ws, err := websocket.Dial("ws://localhost:9222/ws", "", "http://localhost")
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer ws.Close()

	fmt.Println("Connected to ws://localhost:9222/ws")

	// Receive handshake ack
	var msg map[string]interface{}
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		log.Fatal("Receive error:", err)
	}
	fmt.Printf("Received: %+v\n", msg)

	// Send handshake
	handshake := map[string]interface{}{
		"version": "1.0.0",
		"type":    "handshake",
		"id":      "test-1",
		"payload": map[string]string{"client_id": "test-client"},
	}
	if err := websocket.JSON.Send(ws, handshake); err != nil {
		log.Fatal("Send error:", err)
	}
	fmt.Println("Sent handshake")

	// Receive response
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		log.Fatal("Receive error:", err)
	}
	fmt.Printf("Received: %+v\n", msg)

	// Send get_range request
	getRange := map[string]interface{}{
		"version": "1.0.0",
		"type":    "get_range",
		"id":      "test-2",
		"payload": map[string]int{"from": 0, "to": 100},
	}
	if err := websocket.JSON.Send(ws, getRange); err != nil {
		log.Fatal("Send error:", err)
	}
	fmt.Println("Sent get_range")

	// Receive response
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		log.Fatal("Receive error:", err)
	}

	fmt.Printf("\n=== Response ===\n")
	jsonBytes, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Println(string(jsonBytes))

	// Check payload
	if payload, ok := msg["payload"].(map[string]interface{}); ok {
		if frames, ok := payload["frames"].([]interface{}); ok {
			fmt.Printf("\nFrames count: %d\n", len(frames))
		}
	}
}
