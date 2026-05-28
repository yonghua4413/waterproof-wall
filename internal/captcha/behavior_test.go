package captcha

import (
	"testing"

	"local/captcha-service/internal/model"
)

func TestEvaluateTrackAcceptsHumanLikeTrack(t *testing.T) {
	track := []model.TrackPoint{
		{X: 0, Y: 0, T: 0},
		{X: 3, Y: 0, T: 60},
		{X: 10, Y: 1, T: 135},
		{X: 24, Y: 2, T: 220},
		{X: 45, Y: 1, T: 330},
		{X: 69, Y: 3, T: 455},
		{X: 93, Y: 2, T: 580},
		{X: 111, Y: 4, T: 700},
		{X: 120, Y: 2, T: 820},
		{X: 118, Y: 1, T: 900},
		{X: 120, Y: 0, T: 980},
	}
	got := EvaluateTrack(track, 120, 122)
	if !got.OK {
		t.Fatalf("expected ok, got %#v", got)
	}
}

func TestEvaluateTrackRejectsLinearBotTrack(t *testing.T) {
	track := []model.TrackPoint{
		{X: 0, Y: 0, T: 0},
		{X: 15, Y: 0, T: 100},
		{X: 30, Y: 0, T: 200},
		{X: 45, Y: 0, T: 300},
		{X: 60, Y: 0, T: 400},
		{X: 75, Y: 0, T: 500},
		{X: 90, Y: 0, T: 600},
		{X: 105, Y: 0, T: 700},
		{X: 120, Y: 0, T: 800},
	}
	got := EvaluateTrack(track, 120, 120)
	if got.OK {
		t.Fatalf("expected reject, got %#v", got)
	}
}
