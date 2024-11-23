package request

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal"
)

func WriteHash(req *http.Request, payload []byte, key string) error {
	h := sha256.New()
	if _, err := h.Write(payload); err != nil {
		return err
	}
	if _, err := h.Write([]byte(key)); err != nil {
		return err
	}
	data := h.Sum(nil)
	req.Header.Add(internal.HashHeader, hex.EncodeToString(data))
	return nil
}
