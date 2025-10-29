# API Specification

## Versioning
- URL path: `/v1/...`
- Protobuf: add new fields only, never rename or remove
- WebSocket handshake includes `"version":"1.0"` frame
- Sunset header: `Deprecation: true` (date 6 months ahead)

## Base URL
`https://api.myapp.com`

## REST Endpoints

### Auth
POST /v1/auth/register  
Body: `{email, password}`  
201: `{accessToken, refreshToken}`  
400: field errors

POST /v1/auth/login  
Body: `{email, password}`  
200: same as register

POST /v1/auth/refresh  
Cookie: refreshToken  
200: `{accessToken}`

### Chats
GET  /v1/chats  
Header: Authorization Bearer <accessToken>  
200: `[{id, type, members[], lastMessage}]`

POST /v1/chats  
Body: `{type, memberIds[]}`  
201: `{chatId}`

POST /v1/chats/:id/invite  
Body: `{userId}`  
204

### Health
GET  /v1/health  
200: `{status:ok}`

## WebSocket Sub-Protocol
Connect: `wss://api.myapp.com/v1/ws?token=<accessToken>`

Frame format: protobuf (text fallback JSON)

## Protobuf Schemas (pb/api.proto)
```proto
syntax = "proto3";
package v1;

message SendMessage {
  string uuid   = 1; // client-generated
  int64  chatId = 2;
  string body   = 3;
}
message Delivered {
  string uuid = 1;
  int64  msgId= 2;
}
message Read {
  int64 chatId = 1;
  int64 msgId  = 2;
}
message Presence {
  int64 userId = 1;
  bool  online = 2;
}
```

## Error Codes
`400` bad request  
`401` unauthorized  
`403` forbidden  
`404` not found  
`410` gone (deprecated endpoint)  
WS close codes: `1000` normal, `4000` token expired, `4001` rate limited
