package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"local/captcha-service/internal/captcha"
	"local/captcha-service/internal/config"
	"local/captcha-service/internal/model"
	"local/captcha-service/internal/platform"
	"local/captcha-service/internal/store"
	"local/captcha-service/internal/ticket"
)

type Server struct {
	cfg      config.Config
	store    store.Store
	platform platform.Store
}

func NewRouter(cfg config.Config, st store.Store, platformStore ...platform.Store) http.Handler {
	var ps platform.Store
	if len(platformStore) > 0 {
		ps = platformStore[0]
	}
	s := &Server{cfg: cfg, store: st, platform: ps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("POST /api/v1/users/register", s.registerUser)
	mux.HandleFunc("POST /api/v1/users/login", s.loginUser)
	mux.HandleFunc("GET /api/v1/users/me", s.getCurrentUser)
	mux.HandleFunc("GET /api/v1/apps", s.listApps)
	mux.HandleFunc("POST /api/v1/apps", s.createApp)
	mux.HandleFunc("POST /api/v1/apps/rotate-secret", s.rotateAppSecret)
	mux.HandleFunc("POST /api/v1/apps/update-domains", s.updateAppDomains)
	mux.HandleFunc("POST /api/v1/challenge", s.challenge)
	mux.HandleFunc("POST /api/v1/verify", s.verify)
	mux.HandleFunc("POST /api/v1/ticket/check", s.checkTicket)
	mux.HandleFunc("GET /api/v1/apps/usage", s.appUsage)
	mux.HandleFunc("GET /api/demo/config", s.demoConfig)
	mux.HandleFunc("POST /api/demo/verify-captcha", s.demoVerifyCaptcha)

	return s.cors(s.recover(mux))
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !validEmail(req.Email) || len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "invalid_user")
		return
	}
	user, err := s.platform.RegisterUser(r.Context(), req.Email, req.Password)
	if errors.Is(err, platform.ErrConflict) {
		writeError(w, http.StatusConflict, "user_exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user_create_failed")
		return
	}
	token, err := platform.NewUserToken(s.cfg.Secret, user.ID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "token": token, "user": publicUser(user)})
}

func (s *Server) loginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	user, err := s.platform.AuthenticateUser(r.Context(), req.Email, req.Password)
	if errors.Is(err, platform.ErrUnauthorized) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login_failed")
		return
	}
	token, err := platform.NewUserToken(s.cfg.Secret, user.ID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "token": token, "user": publicUser(user)})
}

func (s *Server) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	user, err := s.platform.GetUser(r.Context(), userID)
	if errors.Is(err, platform.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user_lookup_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "user": publicUser(user)})
}

func (s *Server) createApp(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Name    string   `json:"name"`
		Domains []string `json:"domains"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	app, err := s.platform.CreateApp(r.Context(), userID, req.Name, req.Domains)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "app_create_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "app": publicApp(app, true)})
}

func (s *Server) listApps(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	apps, err := s.platform.ListApps(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "app_list_failed")
		return
	}
	out := make([]map[string]any, 0, len(apps))
	for _, app := range apps {
		out = append(out, publicApp(app, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "apps": out})
}

func (s *Server) rotateAppSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		AppID string `json:"appId"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	app, err := s.platform.RotateAppSecret(r.Context(), userID, req.AppID)
	if errors.Is(err, platform.ErrNotFound) {
		writeError(w, http.StatusNotFound, "app_not_found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "secret_rotate_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "app": publicApp(app, true)})
}

func (s *Server) updateAppDomains(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		AppID   string   `json:"appId"`
		Domains []string `json:"domains"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.AppID == "" {
		writeError(w, http.StatusBadRequest, "missing_app_id")
		return
	}
	app, err := s.platform.UpdateAppDomains(r.Context(), userID, req.AppID, req.Domains)
	if errors.Is(err, platform.ErrNotFound) {
		writeError(w, http.StatusNotFound, "app_not_found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update_domains_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "app": publicApp(app, false)})
}

func (s *Server) challenge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID string `json:"appId"`
		Scene string `json:"scene"`
		BizID string `json:"bizId"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Scene = normalize(req.Scene, "default")
	app, ok := s.loadRequestApp(w, r, req.AppID, false)
	if !ok {
		return
	}
	if !platform.AllowedOrigin(app, r.Header.Get("Origin")) {
		writeError(w, http.StatusForbidden, "origin_not_allowed")
		return
	}
	if !s.allow(w, r, app.ID+":challenge:"+req.Scene, s.cfg.ChallengeLimit) {
		return
	}

	assets, err := captcha.Generate(s.cfg, app.ID, req.Scene, req.BizID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generate_failed")
		return
	}
	if err := s.store.SaveChallenge(r.Context(), assets.Challenge, s.cfg.ChallengeTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "store_failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"captchaId":  assets.Challenge.ID,
		"background": assets.BackgroundDataURL,
		"piece":      assets.PieceDataURL,
		"pieceY":     assets.Challenge.PieceY,
		"width":      assets.Challenge.Width,
		"height":     assets.Challenge.Height,
		"pieceSize":  assets.Challenge.PieceSize,
		"expiresIn":  int(s.cfg.ChallengeTTL.Seconds()),
	})
	s.logEvent(r, app.ID, req.Scene, req.BizID, platform.EventChallenge, "")
}

func (s *Server) verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID     string             `json:"appId"`
		CaptchaID string             `json:"captchaId"`
		Scene     string             `json:"scene"`
		BizID     string             `json:"bizId"`
		X         int                `json:"x"`
		Track     []model.TrackPoint `json:"track"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Scene = normalize(req.Scene, "default")
	app, ok := s.loadRequestApp(w, r, req.AppID, false)
	if !ok {
		return
	}
	if !s.allow(w, r, app.ID+":verify:"+req.Scene, s.cfg.VerifyLimit) {
		return
	}
	challenge, err := s.store.GetChallenge(r.Context(), req.CaptchaID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "challenge_not_found")
		return
	}
	if challenge.AppID != app.ID || challenge.Scene != req.Scene || challenge.BizID != req.BizID {
		writeError(w, http.StatusBadRequest, "context_mismatch")
		return
	}
	if time.Now().UTC().After(challenge.ExpiresAt) {
		_ = s.store.DeleteChallenge(r.Context(), req.CaptchaID)
		writeError(w, http.StatusBadRequest, "challenge_expired")
		return
	}

	challenge.Attempts++
	if challenge.Attempts > 3 {
		_ = s.store.DeleteChallenge(r.Context(), req.CaptchaID)
		writeError(w, http.StatusTooManyRequests, "too_many_attempts")
		return
	}
	_ = s.store.UpdateChallenge(r.Context(), challenge, time.Until(challenge.ExpiresAt))

	behavior := captcha.EvaluateTrack(req.Track, req.X, challenge.AnswerX)
	if !behavior.OK {
		s.logEvent(r, app.ID, req.Scene, req.BizID, platform.EventFail, behavior.Reason)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"reason":  behavior.Reason,
			"score":   behavior.Score,
		})
		return
	}

	token, payload, err := ticket.Issue(s.cfg.Secret, challenge.AppID, challenge.ID, challenge.Scene, challenge.BizID, s.cfg.TicketTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ticket_failed")
		return
	}
	record := model.TicketRecord{
		ID:        payload.TicketID,
		AppID:     payload.AppID,
		CaptchaID: payload.CaptchaID,
		Scene:     payload.Scene,
		BizID:     payload.BizID,
		ExpiresAt: time.Unix(payload.ExpiresAt, 0).UTC(),
	}
	if err := s.store.SaveTicket(r.Context(), record, s.cfg.TicketTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "ticket_store_failed")
		return
	}
	_ = s.store.DeleteChallenge(r.Context(), challenge.ID)

	s.logEvent(r, app.ID, req.Scene, req.BizID, platform.EventPass, "")

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"ticket":    token,
		"expiresIn": int(s.cfg.TicketTTL.Seconds()),
		"score":     behavior.Score,
	})
}

func (s *Server) checkTicket(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	var req struct {
		AppID  string `json:"appId"`
		Ticket string `json:"ticket"`
		Scene  string `json:"scene"`
		BizID  string `json:"bizId"`
	}
	if !decodeJSON(w, body, &req) {
		return
	}
	req.Scene = normalize(req.Scene, "default")
	app, ok := s.loadRequestApp(w, r, req.AppID, true)
	if !ok {
		return
	}
	if !platform.VerifyRequestSignature(app, r, body) {
		writeError(w, http.StatusUnauthorized, "invalid_signature")
		return
	}

	payload, err := ticket.Parse(s.cfg.Secret, req.Ticket)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_ticket")
		return
	}
	if payload.AppID != app.ID || payload.Scene != req.Scene || payload.BizID != req.BizID {
		writeError(w, http.StatusUnauthorized, "ticket_context_mismatch")
		return
	}

	record, err := s.store.ConsumeTicket(r.Context(), payload.TicketID)
	if errors.Is(err, store.ErrUsed) {
		writeError(w, http.StatusUnauthorized, "ticket_used")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnauthorized, "ticket_not_found")
		return
	}
	if record.AppID != payload.AppID || record.CaptchaID != payload.CaptchaID || record.Scene != payload.Scene || record.BizID != payload.BizID {
		writeError(w, http.StatusUnauthorized, "ticket_record_mismatch")
		return
	}

	s.logEvent(r, app.ID, req.Scene, req.BizID, platform.EventTicket, "")

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"appId":     payload.AppID,
		"captchaId": payload.CaptchaID,
		"scene":     payload.Scene,
		"bizId":     payload.BizID,
		"issuedAt":  payload.IssuedAt,
		"expiresAt": payload.ExpiresAt,
	})
}

func (s *Server) allow(w http.ResponseWriter, r *http.Request, bucket string, limit int) bool {
	key := clientIP(r) + ":" + bucket
	ok, err := s.store.Allow(r.Context(), key, limit, time.Minute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "rate_limit_failed")
		return false
	}
	if !ok {
		writeError(w, http.StatusTooManyRequests, "rate_limited")
		return false
	}
	return true
}

func (s *Server) cors(next http.Handler) http.Handler {
	allowedAny := false
	allowed := map[string]struct{}{}
	for _, origin := range s.cfg.AllowedOrigins {
		if origin == "*" {
			allowedAny = true
		}
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedAny {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if _, ok := allowed[origin]; ok && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Captcha-App-Key,X-Captcha-Timestamp,X-Captcha-Nonce,X-Captcha-Signature")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeError(w, http.StatusInternalServerError, "internal_error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	body, ok := readBody(w, r)
	if !ok {
		return false
	}
	return decodeJSON(w, body, out)
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "empty_body")
		return nil, false
	}
	defer r.Body.Close()
	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(http.MaxBytesReader(w, r.Body, 512*1024)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return nil, false
	}
	return body.Bytes(), true
}

func decodeJSON(w http.ResponseWriter, body []byte, out any) bool {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	claims, err := platform.ParseUserToken(s.cfg.Secret, token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	user, err := s.platform.GetUser(r.Context(), claims.UserID)
	if errors.Is(err, platform.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user_lookup_failed")
		return "", false
	}
	if user.Status != "active" {
		writeError(w, http.StatusForbidden, "user_inactive")
		return "", false
	}
	return claims.UserID, true
}

func (s *Server) loadRequestApp(w http.ResponseWriter, r *http.Request, appID string, fromKey bool) (platform.App, bool) {
	var (
		app platform.App
		err error
	)
	if fromKey {
		appKey := strings.TrimSpace(r.Header.Get("X-Captcha-App-Key"))
		if appKey == "" {
			writeError(w, http.StatusUnauthorized, "missing_app_key")
			return platform.App{}, false
		}
		app, err = s.platform.GetAppByKey(r.Context(), appKey)
	} else {
		appID = strings.TrimSpace(appID)
		if appID == "" {
			writeError(w, http.StatusBadRequest, "missing_app_id")
			return platform.App{}, false
		}
		app, err = s.platform.GetApp(r.Context(), appID)
	}
	if errors.Is(err, platform.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "app_not_found")
		return platform.App{}, false
	}
	if err != nil {
		log.Printf("loadRequestApp failed (fromKey=%v appID=%q appKey=%q): %v", fromKey, appID, r.Header.Get("X-Captcha-App-Key"), err)
		writeError(w, http.StatusInternalServerError, "app_lookup_failed")
		return platform.App{}, false
	}
	if app.Status != "active" {
		writeError(w, http.StatusForbidden, "app_disabled")
		return platform.App{}, false
	}
	if fromKey && strings.TrimSpace(appID) != "" && app.ID != appID {
		writeError(w, http.StatusUnauthorized, "app_key_mismatch")
		return platform.App{}, false
	}
	return app, true
}

func (s *Server) appUsage(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	appID := strings.TrimSpace(r.URL.Query().Get("appId"))
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	apps, err := s.platform.ListApps(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "app_list_failed")
		return
	}
	appMap := map[string]platform.App{}
	for _, app := range apps {
		appMap[app.ID] = app
	}
	if appID != "" {
		if _, ok := appMap[appID]; !ok {
			writeError(w, http.StatusForbidden, "app_not_owned")
			return
		}
		usage, err := s.platform.GetDailyUsage(r.Context(), appID, date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "usage_query_failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "usage": usage})
		return
	}
	type appUsage struct {
		App       platform.App         `json:"app"`
		Daily     platform.DailyUsage  `json:"daily"`
	}
	out := make([]appUsage, 0, len(apps))
	for _, app := range apps {
		usage, _ := s.platform.GetDailyUsage(r.Context(), app.ID, date)
		out = append(out, appUsage{App: app, Daily: usage})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "date": date, "apps": out})
}

const demoAppID = "app_Ovx-y8QZNRKZK5cSbOTo1mP6A4Gp5BNi"

func (s *Server) demoConfig(w http.ResponseWriter, r *http.Request) {
	endpoint := ""
	writeJSON(w, http.StatusOK, map[string]any{
		"appId":    demoAppID,
		"endpoint": endpoint,
	})
}

func (s *Server) demoVerifyCaptcha(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	var req struct {
		AppID  string `json:"appId"`
		Ticket string `json:"ticket"`
		Scene  string `json:"scene"`
		BizID  string `json:"bizId"`
	}
	if !decodeJSON(w, body, &req) {
		return
	}
	req.Scene = normalize(req.Scene, "default")

	app, ok := s.loadRequestApp(w, r, req.AppID, false)
	if !ok {
		return
	}

	payload, err := ticket.Parse(s.cfg.Secret, req.Ticket)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_ticket")
		return
	}
	if payload.AppID != app.ID || payload.Scene != req.Scene || payload.BizID != req.BizID {
		writeError(w, http.StatusUnauthorized, "ticket_context_mismatch")
		return
	}

	record, err := s.store.ConsumeTicket(r.Context(), payload.TicketID)
	if errors.Is(err, store.ErrUsed) {
		writeError(w, http.StatusUnauthorized, "ticket_used")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnauthorized, "ticket_not_found")
		return
	}
	if record.AppID != payload.AppID || record.CaptchaID != payload.CaptchaID || record.Scene != payload.Scene || record.BizID != payload.BizID {
		writeError(w, http.StatusUnauthorized, "ticket_record_mismatch")
		return
	}

	s.logEvent(r, app.ID, req.Scene, req.BizID, platform.EventTicket, "")

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"appId":     payload.AppID,
		"captchaId": payload.CaptchaID,
		"scene":     payload.Scene,
		"bizId":     payload.BizID,
		"issuedAt":  payload.IssuedAt,
		"expiresAt": payload.ExpiresAt,
	})
}

func (s *Server) logEvent(r *http.Request, appID, scene, bizID string, eventType platform.EventType, reason string) {
	ctx := r.Context()
	event := platform.CaptchaEvent{
		AppID:     appID,
		Scene:     scene,
		BizID:     bizID,
		EventType: eventType,
		Reason:    reason,
		IP:        clientIP(r),
	}
	_ = s.platform.LogEvent(ctx, event)
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > 255 {
		return false
	}
	if !strings.Contains(email, "@") {
		return len(email) >= 3
	}
	return strings.Contains(email, ".")
}

func publicUser(user platform.User) map[string]any {
	return map[string]any{
		"id":        user.ID,
		"email":     user.Email,
		"status":    user.Status,
		"createdAt": user.CreatedAt,
	}
}

func publicApp(app platform.App, includeSecret bool) map[string]any {
	out := map[string]any{
		"appId":     app.ID,
		"name":      app.Name,
		"appKey":    app.AppKey,
		"domains":   app.Domains,
		"status":    app.Status,
		"createdAt": app.CreatedAt,
	}
	if includeSecret {
		out["appSecret"] = app.AppSecret
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"success": false, "error": code})
}

func normalize(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if len(value) > 64 {
		return value[:64]
	}
	return value
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > -1 {
		return host[:idx]
	}
	return host
}
