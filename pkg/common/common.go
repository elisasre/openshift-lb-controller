package common

import (
	"os"
)

// SentryEnabled returns bool value is sentry enabled
func SentryEnabled() bool {
	if len(os.Getenv("SENTRY_DSN")) != 0 {
		return true
	}
	return false
}
