package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

func RandomUint16NotIn(exclude []uint16) (uint16, error) {
	excludeSet := make(map[uint16]struct{}, len(exclude))
	for _, v := range exclude {
		excludeSet[v] = struct{}{}
	}

	for {
		var buf [2]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return 0, fmt.Errorf("random read failed: %w", err)
		}
		n := binary.BigEndian.Uint16(buf[:])

		if _, exists := excludeSet[n]; !exists {
			return n, nil
		}
	}
}
