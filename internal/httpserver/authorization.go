package httpserver

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"
)

// CheckAuthorization reports whether the Authorization header value carries
// the expected bearer token. The comparison is done in constant time (on
// fixed-size digests, so the token length does not leak either).
func CheckAuthorization(authorization string, bearer string) bool {
	scheme, token, found := strings.Cut(authorization, " ")
	if !found || !strings.EqualFold(scheme, "bearer") {
		return false
	}
	got := sha256.Sum256([]byte(strings.TrimSpace(token)))
	want := sha256.Sum256([]byte(bearer))
	return subtle.ConstantTimeCompare(got[:], want[:]) == 1
}
