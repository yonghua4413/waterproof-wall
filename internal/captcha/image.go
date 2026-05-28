package captcha

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"time"

	"local/captcha-service/internal/config"
	"local/captcha-service/internal/model"
	"local/captcha-service/internal/secure"
)

type Assets struct {
	Challenge         model.Challenge
	BackgroundDataURL string
	PieceDataURL      string
}

func Generate(cfg config.Config, appID, scene, bizID string) (Assets, error) {
	id, err := secure.ID(24)
	if err != nil {
		return Assets{}, err
	}
	x, err := secure.Int(cfg.PieceSize+18, cfg.ImageWidth-cfg.PieceSize-18)
	if err != nil {
		return Assets{}, err
	}
	y, err := secure.Int(22, cfg.ImageHeight-cfg.PieceSize-18)
	if err != nil {
		return Assets{}, err
	}

	now := time.Now().UTC()
	challenge := model.Challenge{
		ID:        id,
		AppID:     appID,
		Scene:     scene,
		BizID:     bizID,
		AnswerX:   x,
		PieceY:    y,
		Width:     cfg.ImageWidth,
		Height:    cfg.ImageHeight,
		PieceSize: cfg.PieceSize,
		CreatedAt: now,
		ExpiresAt: now.Add(cfg.ChallengeTTL),
	}

	bg := drawBackground(cfg.ImageWidth, cfg.ImageHeight)
	mask := puzzleMask(cfg.PieceSize)
	piece := image.NewNRGBA(image.Rect(0, 0, cfg.PieceSize, cfg.PieceSize))

	for py := 0; py < cfg.PieceSize; py++ {
		for px := 0; px < cfg.PieceSize; px++ {
			if !mask[py][px] {
				continue
			}
			source := bg.NRGBAAt(x+px, y+py)
			piece.SetNRGBA(px, py, color.NRGBA{R: source.R, G: source.G, B: source.B, A: 255})

			// Carve a visible but non-destructive gap in the background image.
			gray := uint8((uint16(source.R) + uint16(source.G) + uint16(source.B)) / 3)
			bg.SetNRGBA(x+px, y+py, color.NRGBA{
				R: blend(gray, 245, 0.70),
				G: blend(gray, 245, 0.70),
				B: blend(gray, 245, 0.70),
				A: 255,
			})
		}
	}

	addPieceBorder(piece, mask)
	addGapBorder(bg, mask, x, y)

	bgURL, err := pngDataURL(bg)
	if err != nil {
		return Assets{}, err
	}
	pieceURL, err := pngDataURL(piece)
	if err != nil {
		return Assets{}, err
	}

	return Assets{Challenge: challenge, BackgroundDataURL: bgURL, PieceDataURL: pieceURL}, nil
}

func drawBackground(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	seed, _ := secure.Int(1, math.MaxInt32)
	r := rand.New(rand.NewSource(int64(seed)))

	baseA := color.NRGBA{R: uint8(80 + r.Intn(80)), G: uint8(120 + r.Intn(70)), B: uint8(130 + r.Intn(80)), A: 255}
	baseB := color.NRGBA{R: uint8(170 + r.Intn(50)), G: uint8(150 + r.Intn(70)), B: uint8(100 + r.Intn(100)), A: 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t := float64(x+y) / float64(w+h)
			noise := r.Intn(18) - 9
			img.SetNRGBA(x, y, color.NRGBA{
				R: clamp(lerp(baseA.R, baseB.R, t) + noise),
				G: clamp(lerp(baseA.G, baseB.G, t) + noise),
				B: clamp(lerp(baseA.B, baseB.B, t) + noise),
				A: 255,
			})
		}
	}

	for i := 0; i < 24; i++ {
		cx := r.Intn(w)
		cy := r.Intn(h)
		radius := 8 + r.Intn(26)
		col := color.NRGBA{
			R: uint8(80 + r.Intn(150)),
			G: uint8(80 + r.Intn(150)),
			B: uint8(80 + r.Intn(150)),
			A: uint8(26 + r.Intn(45)),
		}
		fillCircle(img, cx, cy, radius, col)
	}

	for i := 0; i < 10; i++ {
		x1, y1 := r.Intn(w), r.Intn(h)
		x2, y2 := r.Intn(w), r.Intn(h)
		drawLine(img, x1, y1, x2, y2, color.NRGBA{R: 255, G: 255, B: 255, A: uint8(25 + r.Intn(45))})
	}

	return img
}

func puzzleMask(size int) [][]bool {
	mask := make([][]bool, size)
	for i := range mask {
		mask[i] = make([]bool, size)
	}
	pad := float64(size) * 0.18
	r := float64(size) * 0.16
	center := float64(size) / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x), float64(y)
			body := fx >= pad && fx <= float64(size)-pad && fy >= pad && fy <= float64(size)-pad
			rightTab := dist(fx, fy, float64(size)-pad, center) <= r
			bottomTab := dist(fx, fy, center, float64(size)-pad) <= r
			topCut := dist(fx, fy, center, pad) <= r
			leftCut := dist(fx, fy, pad, center) <= r
			mask[y][x] = (body || rightTab || bottomTab) && !topCut && !leftCut
		}
	}
	return mask
}

func addPieceBorder(img *image.NRGBA, mask [][]bool) {
	size := len(mask)
	for y := 1; y < size-1; y++ {
		for x := 1; x < size-1; x++ {
			if !mask[y][x] || (mask[y-1][x] && mask[y+1][x] && mask[y][x-1] && mask[y][x+1]) {
				continue
			}
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 230})
		}
	}
}

func addGapBorder(img *image.NRGBA, mask [][]bool, ox, oy int) {
	size := len(mask)
	for y := 1; y < size-1; y++ {
		for x := 1; x < size-1; x++ {
			if !mask[y][x] || (mask[y-1][x] && mask[y+1][x] && mask[y][x-1] && mask[y][x+1]) {
				continue
			}
			img.SetNRGBA(ox+x, oy+y, color.NRGBA{R: 70, G: 85, B: 90, A: 255})
		}
	}
}

func pngDataURL(img image.Image) (string, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func fillCircle(img draw.Image, cx, cy, radius int, c color.NRGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
				continue
			}
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= radius*radius {
				blendPixel(img, x, y, c)
			}
		}
	}
}

func drawLine(img draw.Image, x1, y1, x2, y2 int, c color.NRGBA) {
	dx := int(math.Abs(float64(x2 - x1)))
	dy := -int(math.Abs(float64(y2 - y1)))
	sx, sy := -1, -1
	if x1 < x2 {
		sx = 1
	}
	if y1 < y2 {
		sy = 1
	}
	err := dx + dy
	for {
		blendPixel(img, x1, y1, c)
		if x1 == x2 && y1 == y2 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

func blendPixel(img draw.Image, x, y int, c color.NRGBA) {
	old := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	a := float64(c.A) / 255
	img.Set(x, y, color.NRGBA{
		R: uint8(float64(old.R)*(1-a) + float64(c.R)*a),
		G: uint8(float64(old.G)*(1-a) + float64(c.G)*a),
		B: uint8(float64(old.B)*(1-a) + float64(c.B)*a),
		A: 255,
	})
}

func lerp(a, b uint8, t float64) int {
	return int(float64(a)*(1-t) + float64(b)*t)
}

func blend(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1-t) + float64(b)*t)
}

func clamp(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x1-x2, y1-y2)
}
