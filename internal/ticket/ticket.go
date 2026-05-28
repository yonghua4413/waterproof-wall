package ticket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"local/captcha-service/internal/secure"
)

var (
	ErrMalformed = errors.New("malformed ticket")
	ErrSignature = errors.New("invalid signature")
	ErrExpired   = errors.New("ticket expired")
)

type Payload struct {
	TicketID  string `json:"tid"`
	AppID     string `json:"appId,omitempty"`
	CaptchaID string `json:"cid"`
	Scene     string `json:"scene"`
	BizID     string `json:"bizId"`
	Nonce     string `json:"nonce"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func Issue(secret []byte, appID, captchaID, scene, bizID string, ttl time.Duration) (string, Payload, error) {
	tid, err := secure.ID(24)
	if err != nil {
		return "", Payload{}, err
	}
	nonce, err := secure.ID(16)
	if err != nil {
		return "", Payload{}, err
	}
	now := time.Now().UTC()
	payload := Payload{
		TicketID:  tid,
		AppID:     appID,
		CaptchaID: captchaID,
		Scene:     scene,
		BizID:     bizID,
		Nonce:     nonce,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}

	header := map[string]string{"alg": "HS256", "typ": "captcha-ticket"}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	encodedHeader := base64.RawURLEncoding.EncodeToString(hb)
	encodedPayload := base64.RawURLEncoding.EncodeToString(pb)
	signingInput := encodedHeader + "." + encodedPayload
	signature := sign(secret, signingInput)
	return signingInput + "." + signature, payload, nil
}

func Parse(secret []byte, token string) (Payload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Payload{}, ErrMalformed
	}
	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(sign(secret, signingInput)), []byte(parts[2])) {
		return Payload{}, ErrSignature
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Payload{}, ErrMalformed
	}
	var payload Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		return Payload{}, ErrMalformed
	}
	if time.Now().UTC().Unix() > payload.ExpiresAt {
		return Payload{}, ErrExpired
	}
	if payload.Nonce == "" {
		return Payload{}, ErrMalformed
	}
	return payload, nil
}

func sign(secret []byte, input string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}