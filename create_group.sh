#!/bin/bash

TOKEN=$(cat /tmp/login.json | jq -r .accessToken)

curl -s -X POST http://localhost:8080/v1/chats \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":2,"memberIds":[1,2,3]}' | jq .
