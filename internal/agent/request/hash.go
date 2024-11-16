package request

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
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
	req.Header.Add("HashSHA256", hex.EncodeToString(data))
	return nil
}
