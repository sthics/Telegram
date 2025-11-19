package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseURL = "http://localhost:8080/v1"
	wsURL   = "ws://localhost:8080/v1/ws"
)

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type RegisterResponse struct {
	UserID      float64 `json:"userId"`
	AccessToken string  `json:"accessToken"`
}

type CreateChatResponse struct {
	ChatID int64 `json:"chatId"`
}

func main() {
	fmt.Println("Starting E2E Test...")

	// 1. Register Users
	adminToken, adminID := registerUser("admin")
	memberToken, memberID := registerUser("member")

	fmt.Printf("Admin ID: %d, Member ID: %d\n", adminID, memberID)

	// 2. Create Group Chat
	fmt.Println("\n[Test] Admin creating group chat...")
	chatID := createChat(adminToken, []int64{memberID})
	fmt.Printf("Chat created with ID: %d\n", chatID)

	// 3. Connect WebSockets
	fmt.Println("\n[Test] Connecting WebSockets...")
	adminWS := connectWS(adminToken)
	defer adminWS.Close()
	memberWS := connectWS(memberToken)
	defer memberWS.Close()

	// Start reading from Member WS
	memberMsgs := make(chan map[string]any, 10)
	go readMessages(memberWS, memberMsgs)

	// 4. Test Typing Indicator
	fmt.Println("\n[Test] Admin sending 'Typing' event...")
	sendTyping(adminWS, chatID)

	// Verify Member received Typing
	select {
	case msg := <-memberMsgs:
		if msg["type"] == "Typing" && int64(msg["chatId"].(float64)) == chatID && int64(msg["userId"].(float64)) == adminID {
			fmt.Println("✅ Member received Typing event")
		} else {
			panic(fmt.Sprintf("Unexpected message: %v", msg))
		}
	case <-time.After(5 * time.Second):
		panic("Timeout waiting for Typing event")
	}

	// 5. Test Message Delivery
	fmt.Println("\n[Test] Admin sending 'Hello' message...")
	sendMessage(adminWS, chatID, "Hello World")

	// Verify Member received Message
	select {
	case msg := <-memberMsgs:
		if msg["type"] == "Message" && int64(msg["chatId"].(float64)) == chatID && msg["body"] == "Hello World" {
			fmt.Println("✅ Member received Message")
		} else {
			// Might receive presence or other events, loop?
			// For now, strict check.
			panic(fmt.Sprintf("Unexpected message: %v", msg))
		}
	case <-time.After(5 * time.Second):
		panic("Timeout waiting for Message")
	}

	fmt.Println("\n✅ E2E Test Completed Successfully!")
}

func registerUser(prefix string) (string, int64) {
	email := fmt.Sprintf("%s_%d@test.com", prefix, time.Now().UnixNano())
	password := "password123"
	
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Register failed: %s", string(b)))
	}

	var res RegisterResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.AccessToken, int64(res.UserID)
}

func createChat(token string, memberIDs []int64) int64 {
	body, _ := json.Marshal(map[string]interface{}{
		"type":      2,
		"memberIds": memberIDs,
	})

	req, _ := http.NewRequest("POST", baseURL+"/chats", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Create chat failed: %s", string(b)))
	}

	var res CreateChatResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.ChatID
}

func connectWS(token string) *websocket.Conn {
	header := http.Header{}
	// Note: Gateway expects Authorization header for WebSocket upgrade if using protected route
	// But standard JS WebSocket API doesn't support headers.
	// Our gateway implementation in server.go uses:
	// protected.GET("/ws", wsMiddleware, s.websocketHandler)
	// protected group uses s.authService.JWTMiddleware()
	// The JWT middleware likely checks Authorization header.
	// Go websocket client supports headers.
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		panic(fmt.Sprintf("WebSocket connection failed: %v", err))
	}
	return conn
}

func readMessages(conn *websocket.Conn, ch chan<- map[string]any) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		
		var msg map[string]any
		if err := json.Unmarshal(message, &msg); err == nil {
			// Filter out Presence events for cleaner test output
			if msg["type"] != "Presence" && msg["type"] != "Pong" {
				ch <- msg
			}
		}
	}
}

func sendTyping(conn *websocket.Conn, chatID int64) {
	msg := map[string]any{
		"type":   "Typing",
		"chatId": chatID,
	}
	conn.WriteJSON(msg)
}

func sendMessage(conn *websocket.Conn, chatID int64, text string) {
	msg := map[string]any{
		"type":   "SendMessage",
		"uuid":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"chatId": chatID,
		"body":   text,
	}
	conn.WriteJSON(msg)
}
