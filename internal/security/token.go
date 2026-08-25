package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"t117/internal/domain"
)

type TokenCodec struct {
	secret []byte
	ttl    time.Duration
}
type claims struct {
	UserID domain.ID
	Name   string
	Expiry time.Time
}

func NewTokenCodec(
	secret string,
) TokenCodec {
	return TokenCodec{secret: []byte(secret), ttl: 24 * time.Hour}
}
func (c TokenCodec) Issue(id domain.ID, name string) (string, error) {
	body := fmt.Sprintf("%s|%s|%d", id, name, time.Now().Add(c.ttl).Unix())
	signature := c.sign(body)
	return base64.RawURLEncoding.EncodeToString([]byte(body + "|" + signature)), nil
}
func (c TokenCodec) Parse(value string) (claims, error) {
	decoded, err :=
		base64.RawURLEncoding.
			DecodeString(value)
	if err != nil {
		return claims{}, domain.ErrUnauthorized
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 4 {
		return claims{}, domain.ErrUnauthorized
	}
	body := strings.Join(parts[:3], "|")
	if !hmac.Equal([]byte(parts[3]), []byte(c.sign(body))) {
		return claims{}, domain.ErrUnauthorized
	}
	var timestamp int64
	if _, err := fmt.Sscan(parts[2], &timestamp); err != nil || time.Unix(timestamp, 0).Before(time.Now()) {
		return claims{}, domain.ErrUnauthorized
	}
	return claims{UserID: domain.ID(parts[0]), Name: parts[1], Expiry: time.Unix(timestamp, 0)}, nil
}
func (c TokenCodec) sign(value string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
