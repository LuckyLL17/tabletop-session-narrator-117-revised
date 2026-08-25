package ids

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func New(
	prefix string,
) string {
	raw := make([]byte, 7)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixNano(), hex.EncodeToString(raw))
}
