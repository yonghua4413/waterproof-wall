package secure

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"math/big"
)

func ID(n int) (string, error) {
	b := make([]byte, n)
	if err := Bytes(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Bytes(b []byte) error {
	_, err := io.ReadFull(rand.Reader, b)
	return err
}

func Int(min, max int) (int, error) {
	if max <= min {
		return min, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}
