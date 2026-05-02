package security

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// ============ Panic 拦截 ============

// Recovery returns a Gin middleware that recovers from panics.
// The panic error is written to stderr/audit log, and the client receives only a generic error.
// Stack traces never reach the client.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Write to stderr/audit log (not to client)
				fmt.Fprintf(os.Stderr,
					"[PANIC] method=%s path=%s error=%v\nstack=%s\n",
					c.Request.Method, c.Request.URL.Path, r, string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    9000,
					"message": "系统繁忙，请稍后重试",
				})
			}
		}()
		c.Next()
	}
}

// ============ 日志脱敏 ============

// Logger is a minimal structured logger interface.
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// NopLogger is a no-op logger (falls back to fmt if no real logger provided).
type NopLogger struct{}

func (NopLogger) Infof(format string, args ...interface{})  { fmt.Printf("[INFO] "+format+"\n", args...) }
func (NopLogger) Warnf(format string, args ...interface{})  { fmt.Printf("[WARN] "+format+"\n", args...) }
func (NopLogger) Errorf(format string, args ...interface{}) { fmt.Printf("[ERROR] "+format+"\n", args...) }
func (NopLogger) Debugf(format string, args ...interface{}) {}

// DefaultLogger used when no logger is injected.
var DefaultLogger = NopLogger{}

// RequestLog returns a Gin middleware that logs HTTP requests.
// Sensitive fields (password, mobile, token, etc.) are automatically masked.
func RequestLog(logger Logger) gin.HandlerFunc {
	if logger == nil {
		logger = DefaultLogger
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Next()
		status := c.Writer.Status()
		// Log without sensitive query/body details
		logLine := fmt.Sprintf("[HTTP] %s %s %d %s",
			method, path, status, c.ClientIP())
		if status >= 500 {
			logger.Errorf("%s", logLine)
		} else if status >= 400 {
			logger.Warnf("%s", logLine)
		} else {
			logger.Infof("%s", logLine)
		}
	}
}

// ============ 错误响应安全封装 ============

// ErrResponse is the only error structure returned to clients.
type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InternalErr returns a generic internal error. Use this instead of returning raw errors.
func InternalErr() ErrResponse {
	return ErrResponse{Code: 9000, Message: "系统繁忙，请稍后重试"}
}

// SafeErr maps internal errors to generic user-facing responses.
// Internal details (SQL errors, Redis keys, file paths) are NEVER returned to the client.
func SafeErr(err error) ErrResponse {
	if err == nil {
		return InternalErr()
	}
	msg := err.Error()
	switch {
	case containsAny(msg, "not found", "record not found", "no rows"):
		return ErrResponse{Code: 7002, Message: "资源不存在"}
	case containsAny(msg, "unauthorized", "permission denied", "forbidden"):
		return ErrResponse{Code: 7003, Message: "无权限操作"}
	case containsAny(msg, "timeout", "context deadline"):
		return ErrResponse{Code: 9003, Message: "请求超时，请稍后重试"}
	case containsAny(msg, "duplicate", "unique constraint", "already exists"):
		return ErrResponse{Code: 7001, Message: "数据已存在，请勿重复提交"}
	case containsAny(msg, "validation", "invalid", "malformed"):
		return ErrResponse{Code: 7001, Message: "请求参数有误"}
	default:
		return InternalErr()
	}
}

func containsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// ============ 敏感字段脱敏 ============

var sensitiveKeyPatterns = []string{
	"password", "pwd", "passwd", "pay_password", "paypwd",
	"token", "access_token", "refresh_token", "jwt", "authorization",
	"mobile", "phone", "tel", "id_card", "idcard", "card_no", "bank_card",
	"email", "secret", "api_key", "apikey",
}

var maskRe = regexp.MustCompile(`"([^"]+)"`)

// MaskFields takes a map and returns a new map with sensitive fields masked.
// Use this before logging or serializing any request/response data.
func MaskFields(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			result[k] = maskValue(v)
		} else if nested, ok := v.(map[string]interface{}); ok {
			result[k] = MaskFields(nested)
		} else {
			result[k] = v
		}
	}
	return result
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, pattern := range sensitiveKeyPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func maskValue(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	if s == "" || s == "<nil>" {
		return "****"
	}
	// 手机号 11位：前3后4
	if matched, _ := regexp.MatchString(`^\d{11}$`, s); matched {
		return s[:3] + "****" + s[7:]
	}
	// 银行卡/证件：前4后4
	if len(s) >= 8 {
		return s[:4] + "****" + s[len(s)-4:]
	}
	// 通用：保留首尾各1字符
	if len(s) > 4 {
		return s[:2] + "****" + s[len(s)-2:]
	}
	return "****"
}

// MaskJSON takes a JSON-like string and masks sensitive values.
// Returns the masked string (does not parse/serialize — lightweight mask for log lines).
func MaskJSON(s string) string {
	if s == "" {
		return s
	}
	for _, pattern := range sensitiveKeyPatterns {
		re := regexp.MustCompile(`(?i)("` + pattern + `"\s*:\s*)"([^"]+)"`)
		s = re.ReplaceAllStringFunc(s, func(match string) string {
			// Replace value with masked
			re2 := regexp.MustCompile(`"([^"]+)"\s*$`)
			return re2.ReplaceAllString(match, `"****"`)
		})
	}
	return s
}

// SafeFieldValue returns a masked value suitable for logging.
// Use this whenever logging request fields that might contain sensitive data.
func SafeFieldValue(key, value string) string {
	if isSensitiveKey(key) {
		return maskValue(value)
	}
	return value
}
