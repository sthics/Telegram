#!/bin/bash
set -e

# Register User A
EMAIL_A="a_$(date +%s)@test.com"
echo "Registering User A: $EMAIL_A"
TOKEN_A=$(curl -s -X POST http://localhost:8080/v1/auth/register -d '{"email":"'$EMAIL_A'", "password":"password123"}' | jq -r .accessToken)

# Register User B
EMAIL_B="b_$(date +%s)@test.com"
echo "Registering User B: $EMAIL_B"
TOKEN_B_JSON=$(curl -s -X POST http://localhost:8080/v1/auth/register -d '{"email":"'$EMAIL_B'", "password":"password123"}')
TOKEN_B=$(echo $TOKEN_B_JSON | jq -r .accessToken)
ID_B=$(echo $TOKEN_B_JSON | jq -r .userId)

# A creates chat with B
echo "A creating chat with B (ID: $ID_B)"
CHAT_ID=$(curl -s -X POST -H "Authorization: Bearer $TOKEN_A" -H "Content-Type: application/json" -d '{"type":1, "memberIds":['$ID_B']}' http://localhost:8080/v1/chats | jq -r .chatId)
echo "Created Chat: $CHAT_ID"

# B sends message
echo "B sending message..."
curl -s -X POST -H "Authorization: Bearer $TOKEN_B" -H "Content-Type: application/json" -d '{"body":"Hello"}' http://localhost:8080/v1/chats/$CHAT_ID/messages > /dev/null

# A gets chats (should have unread count)
echo "A checking chats..."
curl -s -H "Authorization: Bearer $TOKEN_A" http://localhost:8080/v1/chats | jq .

# A sends message (should NOT increase unread count for A)
echo "A sending message..."
curl -s -X POST -H "Authorization: Bearer $TOKEN_A" -H "Content-Type: application/json" -d '{"body":"Reply"}' http://localhost:8080/v1/chats/$CHAT_ID/messages > /dev/null

# A gets chats again (unread count should still be 1 from B's message, not 2)
echo "A checking chats (expect unread=1)..."
curl -s -H "Authorization: Bearer $TOKEN_A" http://localhost:8080/v1/chats | jq .
