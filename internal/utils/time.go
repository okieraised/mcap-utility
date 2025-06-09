package utils

import (
	"fmt"
	"time"
)

var timestampLayouts = []string{
	time.RFC3339,                    // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05",           // ISO 8601 without timezone
	"2006-01-02 15:04:05",           // Space-separated date/time
	"01/02/2006 15:04:05",           // US format
	"02/01/2006 15:04:05",           // EU format
	"2006-01-02T15:04:05.000Z",      // ISO with milliseconds and Z
	"2006-01-02T15:04:05.999999999", // Full Go time
}

func TryParseTimestamp(input string) (int64, error) {
	for _, layout := range timestampLayouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t.UnixNano(), nil
		}
	}
	return -1, fmt.Errorf("unable to parse timestamp: %s", input)
}
