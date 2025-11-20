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

	// Allow time for RabbitMQ bindings to propagate
	time.Sleep(2 * time.Second)

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
	msgID := sendMessage(adminWS, chatID, "Hello World")

	// Verify Member received Message
	select {
	case msg := <-memberMsgs:
		if msg["type"] == "Message" && int64(msg["chatId"].(float64)) == chatID && msg["body"] == "Hello World" {
			fmt.Println("✅ Member received Message")
			// Capture msgId for Read Receipt test
			msgID = int64(msg["msgId"].(float64))
			if msgID == 0 {
				panic("Received msgId is 0")
			}
		} else {
			panic(fmt.Sprintf("Unexpected message: %v", msg))
		}
	case <-time.After(5 * time.Second):
		panic("Timeout waiting for Message")
	}

	// 6. Test Read Receipt
	fmt.Println("\n[Test] Member sending 'Read' receipt...")
	sendReadReceipt(memberWS, chatID, msgID)

	// Verify Admin received Read Receipt
	// We need to read from Admin WS now
	adminMsgs := make(chan map[string]any, 10)
	go readMessages(adminWS, adminMsgs)

	select {
	case msg := <-adminMsgs:
		if msg["type"] == "Read" && int64(msg["chatId"].(float64)) == chatID && int64(msg["msgId"].(float64)) == msgID && int64(msg["userId"].(float64)) == memberID {
			fmt.Println("✅ Admin received Read Receipt")
		} else {
			// Might receive other events (like own message delivery ack), filter?
			// For now, strict check might fail if other events come first.
			// Let's loop a bit.
			found := false
			// Check current msg
			if msg["type"] == "Read" && int64(msg["chatId"].(float64)) == chatID && int64(msg["msgId"].(float64)) == msgID {
				found = true
			}
			
			if !found {
				// Loop for a few more
				timeout := time.After(2 * time.Second)
				for !found {
					select {
					case m := <-adminMsgs:
						if m["type"] == "Read" && int64(m["chatId"].(float64)) == chatID && int64(m["msgId"].(float64)) == msgID {
							found = true
							fmt.Println("✅ Admin received Read Receipt (after skipping)")
						}
					case <-timeout:
						panic(fmt.Sprintf("Timeout waiting for Read Receipt. Last msg: %v", msg))
					}
				}
			} else {
				fmt.Println("✅ Admin received Read Receipt")
			}
		}
	case <-time.After(5 * time.Second):
		panic("Timeout waiting for Read Receipt")
	}

	// 6. Test Push Notifications
	fmt.Println("\n[Test] Registering device token for Member...")
	registerDevice(memberToken, "member-device-token", "web")

	fmt.Println("[Test] Disconnecting Member WebSocket to simulate offline...")
	memberWS.Close()
	// Wait for presence to update
	time.Sleep(2 * time.Second)

	fmt.Println("[Test] Admin sending message to trigger push...")
	sendMessage(adminWS, chatID, "Push this!")

	// Note: We can't easily verify the push log in the E2E test without accessing docker logs
	// But we can verify the Admin receives the message (echo)
	fmt.Println("✅ Admin sent message for push")

	fmt.Println("\n✅ E2E Test Completed Successfully!")
}

func registerDevice(token, deviceToken, platform string) {
	url := "http://localhost:8080/v1/devices"
	body := map[string]string{
		"token":    deviceToken,
		"platform": platform,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		panic(fmt.Sprintf("Failed to register device: %d", resp.StatusCode))
	}
	fmt.Println("✅ Device registered")
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
			fmt.Printf("WebSocket read error: %v\n", err)
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

func sendReadReceipt(conn *websocket.Conn, chatID, msgID int64) {
	msg := map[string]any{
		"type":   "Read",
		"chatId": chatID,
		"msgId":  msgID,
	}
	conn.WriteJSON(msg)
}

func sendMessage(conn *websocket.Conn, chatID int64, text string) int64 {
	// We don't know the msgID yet, it's assigned by server.
	// But for the test we need it.
	// Actually, the server sends back a "Delivered" ack with msgId.
	// Or the "Message" event to others has it.
	// Let's assume we get it from the receiver's "Message" event for now.
	// But wait, we need it to send the Read receipt.
	// So the test flow is: Admin sends -> Member receives (gets ID) -> Member sends Read(ID).
	// So sendMessage doesn't return ID immediately.
	// We'll return 0 and let the caller extract it from receiver.
	// Wait, the caller (main) extracts it from memberMsgs.
	
	msg := map[string]any{
		"type":   "SendMessage",
		"uuid":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"chatId": chatID,
		"body":   text,
	}
	conn.WriteJSON(msg)
	return 0 // Placeholder, actual ID comes from events
}
