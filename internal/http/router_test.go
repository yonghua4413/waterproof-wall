package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local/captcha-service/internal/config"
	"local/captcha-service/internal/model"
	"local/captcha-service/internal/platform"
	"local/captcha-service/internal/store"
)

func TestFullPlatformFlow(t *testing.T) {
	cfg := config.Config{
		Secret:         []byte("test-secret-32-bytes-long-enough"),
		AllowedOrigins: []string{"*"},
		ChallengeLimit: 10,
		VerifyLimit:    10,
		ChallengeTTL:   2 * time.Minute,
		TicketTTL:      2 * time.Minute,
		ImageWidth:     320,
		ImageHeight:    160,
		PieceSize:      58,
	}
	st := store.NewMemory()
	ps := platform.NewMemory()
	handler := NewRouter(cfg, st, ps)

	user, err := ps.RegisterUser(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	app, err := ps.CreateApp(context.Background(), user.ID, "Test App", []string{"*"})
	if err != nil {
		t.Fatal(err)
	}

	challengeRes := doJSON(t, handler, "POST", "/api/v1/challenge", nil, map[string]string{
		"appId": app.ID,
		"scene": "login",
		"bizId": "user-1",
	})
	if challengeRes.Code != http.StatusOK {
		t.Fatalf("challenge status = %d body=%s", challengeRes.Code, challengeRes.Body.String())
	}

	var challengeBody struct {
		CaptchaID string `json:"captchaId"`
	}
	decodeBody(t, challengeRes, &challengeBody)
	challenge, err := st.GetChallenge(context.Background(), challengeBody.CaptchaID)
	if err != nil {
		t.Fatal(err)
	}

	verifyRes := doJSON(t, handler, "POST", "/api/v1/verify", nil, map[string]any{
		"appId":     app.ID,
		"captchaId": challenge.ID,
		"scene":     "login",
		"bizId":     "user-1",
		"x":         challenge.AnswerX,
		"track":     humanTrack(challenge.AnswerX),
	})
	if verifyRes.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", verifyRes.Code, verifyRes.Body.String())
	}

	var verifyBody struct {
		Success bool   `json:"success"`
		Ticket  string `json:"ticket"`
	}
	decodeBody(t, verifyRes, &verifyBody)
	if !verifyBody.Success || verifyBody.Ticket == "" {
		t.Fatalf("verify failed: %s", verifyRes.Body.String())
	}

	body := map[string]string{
		"appId":  app.ID,
		"ticket": verifyBody.Ticket,
		"scene":  "login",
		"bizId":  "user-1",
	}
	bodyBytes, _ := json.Marshal(body)
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := "test-nonce"
	signingInput := fmt.Sprintf("POST\n/api/v1/ticket/check\n%s\n%s\n%s", ts, nonce, string(bodyBytes))
	mac := hmac.New(sha256.New, []byte(app.AppSecret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	ticketHeaders := map[string]string{
		"X-Captcha-App-Key":    app.AppKey,
		"X-Captcha-Timestamp":  ts,
		"X-Captcha-Nonce":      nonce,
		"X-Captcha-Signature":  sig,
	}
	checkRes := doJSONWithBody(t, handler, "POST", "/api/v1/ticket/check", ticketHeaders, bodyBytes)
	if checkRes.Code != http.StatusOK {
		t.Fatalf("ticket check status = %d body=%s", checkRes.Code, checkRes.Body.String())
	}

	replayRes := doJSONWithBody(t, handler, "POST", "/api/v1/ticket/check", ticketHeaders, bodyBytes)
	if replayRes.Code != http.StatusUnauthorized {
		t.Fatalf("replay should fail, got status = %d", replayRes.Code)
	}
}

func doJSON(t *testing.T, handler http.Handler, method, path string, headers map[string]string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return doJSONWithBody(t, handler, method, path, headers, data)
}

func doJSONWithBody(t *testing.T, handler http.Handler, method, path string, headers map[string]string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func decodeBody(t *testing.T, res *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(res.Body.Bytes(), out); err != nil {
		t.Fatalf("decode body: %v body=%s", err, res.Body.String())
	}
}

func humanTrack(x int) []model.TrackPoint {
	return []model.TrackPoint{
		{X: 0, Y: 0, T: 0},
		{X: x / 40, Y: 0, T: 45},
		{X: x / 12, Y: 1, T: 120},
		{X: x / 5, Y: 2, T: 210},
		{X: x / 3, Y: 1, T: 325},
		{X: x / 2, Y: 3, T: 460},
		{X: x*2/3 + 3, Y: 2, T: 615},
		{X: x*4/5 + 4, Y: 4, T: 760},
		{X: x - 7, Y: 2, T: 895},
		{X: x + 2, Y: 1, T: 1030},
		{X: x, Y: 0, T: 1120},
	}
}