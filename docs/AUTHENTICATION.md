# Authentication

## JWT
Algorithm: ES256 ( ECDSA P-256 )  
Access token life: 15 min  
Refresh token life: 7 days (HttpOnly cookie)

## Payload
```json
{
  "sub": 123456,        // userId
  "iss": "minitelegram",
  "exp": 1710000000,
  "iat": 1709999100
}
```

## Middleware (Gin)
```go
func JWT() gin.HandlerFunc {
  return func(c *gin.Context) {
    tokenString := extract(c)
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
      return publicKey, nil
    })
    if err != nil || !token.Valid {
       c.AbortWithStatusJSON(401, gin.H{"code":"INVALID_TOKEN"})
       return
    }
    c.Set("uid", claims.Subject)
    c.Next()
  }
}
```

## Password Hashing
bcrypt cost 12 (≈250 ms on 2 GHz core)

## Refresh Flow
POST /v1/auth/refresh  
Cookie: refreshToken  
Server rotates refresh token (revoke old) → sets new HttpOnly cookie

## Security Checklist
- Tokens signed server-side only (private key in K8s secret)
- Rotation on password change
- Rate-limit login endpoint (see SECURITY_CHECKLIST.md)
