package utils

import (
	"fmt"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	var num int
	var unit string
	if _, err := fmt.Sscanf(s, "%d%s", &num, &unit); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	switch unit {
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s (supported: s, m, h, d)", unit)
	}
}

func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
