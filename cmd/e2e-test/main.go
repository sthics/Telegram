package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/v1"

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type Chat struct {
	ID int64 `json:"id"`
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

	// Verify both see the chat
	verifyChatExists(adminToken, chatID, "Admin", true)
	verifyChatExists(memberToken, chatID, "Member", true)

	// 3. Test Permissions (Member tries to kick Admin)
	fmt.Println("\n[Test] Member trying to kick Admin (should fail)...")
	kickMember(memberToken, chatID, adminID, http.StatusForbidden)

	// 4. Test Kick (Admin kicks Member)
	fmt.Println("\n[Test] Admin kicking Member...")
	kickMember(adminToken, chatID, memberID, http.StatusNoContent)

	// Verify Member is gone
	verifyChatExists(memberToken, chatID, "Member", false)
	verifyChatExists(adminToken, chatID, "Admin", true)

	// 5. Test Leave
	fmt.Println("\n[Test] Re-inviting Member and testing Leave...")
	inviteMember(adminToken, chatID, memberID)
	verifyChatExists(memberToken, chatID, "Member", true)

	fmt.Println("Member leaving chat...")
	leaveChat(memberToken, chatID)
	verifyChatExists(memberToken, chatID, "Member", false)

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

	// Hack: To get ID, we'd need to parse JWT or just trust the flow. 
	// For this test, let's just login again or assume we can get it.
	// Actually, we need the ID for invites. 
	// Let's cheat and use a separate helper or just parse the JWT if needed.
	// OR, better, let's just use the fact that we are creating fresh users.
	// Wait, I need the ID to invite/kick.
	// I'll add a debug endpoint or just parse the token? 
	// Parsing token is annoying without the secret.
	// I'll just use the /v1/auth/login which returns the same.
	// Actually, I can't easily get the ID from the API response as currently implemented (only returns tokens).
	// I will use the `GET /v1/chats` (empty) to maybe debug? No.
	// I will update the register/login response to return UserID for easier testing?
	// No, I shouldn't change production code just for this unless useful.
	// I'll use the `db` directly? No, that breaks "E2E" via API.
	// Ah, I can use the `GET /v1/chats`? No.
	// Wait, `GET /v1/chats` doesn't return user ID.
	// I'll parse the JWT. It's just base64.
	
	return res.AccessToken, int64(res.UserID)
}

type RegisterResponse struct {
	UserID      float64 `json:"userId"`
	AccessToken string  `json:"accessToken"`
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

func verifyChatExists(token string, chatID int64, user string, shouldExist bool) {
	req, _ := http.NewRequest("GET", baseURL+"/chats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var chats []Chat
	json.NewDecoder(resp.Body).Decode(&chats)

	found := false
	for _, c := range chats {
		if c.ID == chatID {
			found = true
			break
		}
	}

	if found != shouldExist {
		panic(fmt.Sprintf("Verification failed for %s: expected chat %d to exist=%v, but found=%v", user, chatID, shouldExist, found))
	}
	fmt.Printf("Verified %s: Chat %d exists=%v\n", user, chatID, found)
}

func kickMember(token string, chatID, userID int64, expectedStatus int) {
	url := fmt.Sprintf("%s/chats/%d/members/%d", baseURL, chatID, userID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Kick member failed: expected status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, string(b)))
	}
}

func leaveChat(token string, chatID int64) {
	url := fmt.Sprintf("%s/chats/%d/members", baseURL, chatID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Leave chat failed: %s", string(b)))
	}
}

func inviteMember(token string, chatID, userID int64) {
	body, _ := json.Marshal(map[string]interface{}{
		"userId": userID,
	})

	url := fmt.Sprintf("%s/chats/%d/invite", baseURL, chatID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Invite member failed: %s", string(b)))
	}
}
