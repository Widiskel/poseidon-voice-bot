package fingerprint

import (
	"crypto/rand"
	"fmt"
	"time"
)

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func randBase62(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {

		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> uint(i*7))
		}
	}
	for i := range b {
		b[i] = base62[int(b[i])%len(base62)]
	}
	return string(b)
}

func MakeRequestID() string {
	return fmt.Sprintf("%d.%s", time.Now().UnixMilli(), randBase62(6))
}
