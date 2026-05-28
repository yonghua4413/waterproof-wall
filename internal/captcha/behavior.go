package captcha

import (
	"math"
	"sort"

	"local/captcha-service/internal/model"
)

type BehaviorResult struct {
	OK     bool    `json:"ok"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

func normalizeTrack(track []model.TrackPoint) []model.TrackPoint {
	if len(track) == 0 {
		return track
	}
	out := make([]model.TrackPoint, len(track))
	out[0] = track[0]
	out[0].T = 1
	for i := 1; i < len(track); i++ {
		out[i] = track[i]
		if out[i].T <= out[i-1].T {
			out[i].T = out[i-1].T + 1
		}
	}
	return out
}

func EvaluateTrack(raw []model.TrackPoint, submittedX, expectedX int) BehaviorResult {
	track := normalizeTrack(raw)
	if len(track) < 8 {
		return BehaviorResult{Reason: "track_too_short"}
	}
	if abs(submittedX-expectedX) > tolerance(expectedX) {
		return BehaviorResult{Reason: "position_mismatch"}
	}

	first := track[0]
	last := track[len(track)-1]
	duration := last.T - first.T
	if duration < 280 {
		return BehaviorResult{Reason: "too_fast"}
	}
	if duration > 12000 {
		return BehaviorResult{Reason: "too_slow"}
	}
	if abs(last.X-submittedX) > 8 {
		return BehaviorResult{Reason: "final_x_mismatch"}
	}

	var reverse, repeated, yMoves int
	var velocities []float64
	var accelerations []float64
	prevV := 0.0
	for i := 1; i < len(track); i++ {
		dx := track[i].X - track[i-1].X
		dy := track[i].Y - track[i-1].Y
		dt := track[i].T - track[i-1].T
		if dx < 0 {
			reverse++
		}
		if dx == 0 && dy == 0 {
			repeated++
		}
		if dy != 0 {
			yMoves++
		}
		v := float64(dx) / float64(dt)
		velocities = append(velocities, v)
		accelerations = append(accelerations, v-prevV)
		prevV = v
	}

	score := 1.0
	if yMoves == 0 {
		score -= 0.18
	}
	if reverse == 0 {
		score -= 0.08
	}
	if repeated > len(track)/3 {
		score -= 0.20
	}
	if variance(velocities) < 0.0008 {
		score -= 0.25
	}
	if variance(accelerations) < 0.0005 {
		score -= 0.15
	}
	if len(uniqueDeltas(track)) <= 4 {
		score -= 0.18
	}
	if percentileStep(track, 0.90) > 45 {
		score -= 0.12
	}

	if score < 0.58 {
		return BehaviorResult{Score: score, Reason: "bot_like_track"}
	}
	return BehaviorResult{OK: true, Score: score}
}

func tolerance(expectedX int) int {
	if expectedX < 120 {
		return 5
	}
	return 6
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))
	total := 0.0
	for _, v := range values {
		d := v - mean
		total += d * d
	}
	return total / float64(len(values))
}

func uniqueDeltas(track []model.TrackPoint) map[int]struct{} {
	out := map[int]struct{}{}
	for i := 1; i < len(track); i++ {
		out[track[i].X-track[i-1].X] = struct{}{}
	}
	return out
}

func percentileStep(track []model.TrackPoint, p float64) int {
	var steps []int
	for i := 1; i < len(track); i++ {
		steps = append(steps, abs(track[i].X-track[i-1].X))
	}
	sort.Ints(steps)
	idx := int(math.Ceil(float64(len(steps))*p)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(steps) {
		idx = len(steps) - 1
	}
	return steps[idx]
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
