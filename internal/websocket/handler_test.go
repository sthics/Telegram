package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ReadPump(t *testing.T) {
	// Setup WebSocket server
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		
		handler := NewHandler(conn, 1, "test-device", zerolog.Nop())
		
		// Start write pump
		go handler.WritePump(time.Second)
		
		// Echo back messages
		handler.ReadPump(func(msg []byte) error {
			return handler.Send(msg)
		})
	}))
	defer server.Close()

	// Connect client
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Send message
	message := []byte(`{"type":"Ping"}`)
	err = conn.WriteMessage(websocket.TextMessage, message)
	assert.NoError(t, err)

	// Read response (echo)
	_, received, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, message, received)
}

func TestHandler_Send(t *testing.T) {
	// Setup WebSocket server
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		
		handler := NewHandler(conn, 1, "test-device", zerolog.Nop())
		
		// Start write pump
		go handler.WritePump(time.Second)
		
		// Send a message
		err = handler.SendJSON(map[string]string{"type": "Test"})
		if err != nil {
			t.Errorf("failed to send: %v", err)
		}
	}))
	defer server.Close()

	// Connect client
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Read message
	_, received, err := conn.ReadMessage()
	assert.NoError(t, err)
	
	var msg map[string]string
	err = json.Unmarshal(received, &msg)
	assert.NoError(t, err)
	assert.Equal(t, "Test", msg["type"])
}
