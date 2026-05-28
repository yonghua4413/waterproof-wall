package model

import "time"

type Challenge struct {
	ID        string    `json:"id"`
	AppID     string    `json:"appId"`
	Scene     string    `json:"scene"`
	BizID     string    `json:"bizId"`
	AnswerX   int       `json:"answerX"`
	PieceY    int       `json:"pieceY"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	PieceSize int       `json:"pieceSize"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Attempts  int       `json:"attempts"`
}

type TrackPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
	T int `json:"t"`
}

type TicketRecord struct {
	ID        string    `json:"id"`
	AppID     string    `json:"appId"`
	CaptchaID string    `json:"captchaId"`
	Scene     string    `json:"scene"`
	BizID     string    `json:"bizId"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `json:"used"`
}
