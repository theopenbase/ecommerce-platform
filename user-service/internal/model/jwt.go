package model

// JWTClaims 是 JWT token 中的 payload 结构（对称 jwt.Manager.Claims）
type JWTClaims struct {
	UserID    uint64 `json:"user_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"
}
