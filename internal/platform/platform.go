package platform

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"local/captcha-service/internal/secure"
)

type EventType string

const (
	EventChallenge EventType = "challenge"
	EventVerify    EventType = "verify"
	EventPass      EventType = "pass"
	EventFail      EventType = "fail"
	EventTicket    EventType = "ticket_check"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidDomain = errors.New("invalid domain")
	ErrInactive      = errors.New("user inactive")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
}

type App struct {
	ID           string
	UserID       string
	Name         string
	AppKey       string
	AppSecret    string
	SecretCipher string
	Domains      []string
	Status       string
	CreatedAt    time.Time
}

type CaptchaEvent struct {
	ID        string
	AppID     string
	Scene     string
	BizID     string
	EventType EventType
	Reason    string
	IP        string
	CreatedAt time.Time
}

type DailyUsage struct {
	AppID     string
	Date      string
	Challenges int64
	Verifies   int64
	Passes     int64
	Fails      int64
	Tickets    int64
}

type Store interface {
	RegisterUser(ctx context.Context, email, password string) (User, error)
	AuthenticateUser(ctx context.Context, email, password string) (User, error)
	ActivateUser(ctx context.Context, userID string) (User, error)
	GetUser(ctx context.Context, userID string) (User, error)
	CreateApp(ctx context.Context, userID, name string, domains []string) (App, error)
	ListApps(ctx context.Context, userID string) ([]App, error)
	GetApp(ctx context.Context, appID string) (App, error)
	GetAppByKey(ctx context.Context, appKey string) (App, error)
	UpdateAppDomains(ctx context.Context, userID, appID string, domains []string) (App, error)
	RotateAppSecret(ctx context.Context, userID, appID string) (App, error)
	LogEvent(ctx context.Context, event CaptchaEvent) error
	GetDailyUsage(ctx context.Context, appID, date string) (DailyUsage, error)
	Close() error
}

type TokenClaims struct {
	UserID    string `json:"uid"`
	ExpiresAt int64  `json:"exp"`
}

func NewUserToken(secret []byte, userID string, ttl time.Duration) (string, error) {
	claims := TokenClaims{UserID: userID, ExpiresAt: time.Now().UTC().Add(ttl).Unix()}
	body, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(body)
	return payload + "." + sign(secret, payload), nil
}

func ParseUserToken(secret []byte, token string) (TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return TokenClaims{}, ErrUnauthorized
	}
	if !hmac.Equal([]byte(sign(secret, parts[0])), []byte(parts[1])) {
		return TokenClaims{}, ErrUnauthorized
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return TokenClaims{}, ErrUnauthorized
	}
	var claims TokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return TokenClaims{}, ErrUnauthorized
	}
	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return TokenClaims{}, ErrUnauthorized
	}
	return claims, nil
}

func NewCredential(prefix string) (string, error) {
	id, err := secure.ID(24)
	if err != nil {
		return "", err
	}
	return prefix + "_" + id, nil
}

func HashPassword(password string) (string, error) {
	salt, err := secure.ID(18)
	if err != nil {
		return "", err
	}
	sum := passwordDigest(salt, password)
	return salt + ":" + base64.RawURLEncoding.EncodeToString(sum), nil
}

func CheckPassword(encoded, password string) bool {
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 {
		return false
	}
	expected, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	got := passwordDigest(parts[0], password)
	return hmac.Equal(got, expected)
}

func VerifyRequestSignature(app App, r *http.Request, body []byte) bool {
	if app.AppSecret == "" {
		return false
	}
	ts := strings.TrimSpace(r.Header.Get("X-Captcha-Timestamp"))
	nonce := strings.TrimSpace(r.Header.Get("X-Captcha-Nonce"))
	sig := strings.TrimSpace(r.Header.Get("X-Captcha-Signature"))
	if ts == "" || nonce == "" || sig == "" {
		return false
	}
	input := strings.Join([]string{r.Method, r.URL.Path, ts, nonce, string(body)}, "\n")
	return hmac.Equal([]byte(sign([]byte(app.AppSecret), input)), []byte(sig))
}

func AllowedOrigin(app App, origin string) bool {
	if len(app.Domains) == 0 || origin == "" {
		return true
	}
	origin = strings.TrimSpace(strings.ToLower(origin))
	for _, domain := range app.Domains {
		domain = strings.TrimSpace(strings.ToLower(domain))
		if domain == "" {
			continue
		}
		if domain == "*" || domain == origin || strings.Contains(origin, "://"+domain) {
			return true
		}
	}
	return false
}

func passwordDigest(salt, password string) []byte {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	out := sum[:]
	for i := 0; i < 60000; i++ {
		next := sha256.Sum256(out)
		out = next[:]
	}
	return out
}

func sign(secret []byte, input string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
