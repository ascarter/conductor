// UUID generator for RFC 4122
// Reference: http://www.ietf.org/rfc/rfc4122.txt

package conductor

import (
	"crypto/rand"
	"fmt"
)

// UUID is representation of RFC 4122 specification
type UUID [16]byte

// NewUUID returns a new UUID string
func NewUUID() (u *UUID) {
	u = new(UUID)
	n, err := rand.Read(u[:])
	if n != len(u) || err != nil {
		return
	}

	// Version 4 (pseudo-random): see RFC 4122 section 4.1.3
	u[6] = (u[6] & 0x0f) | 0x40

	// Variant bits: see RFC 4122 section 4.1.1
	u[8] = (u[8] & 0x3f) | 0x80

	return
}

func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}
