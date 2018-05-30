package common

import (
	"os"
)

func SentryEnabled() bool {
	if len(os.Getenv("SENTRY_DSN")) != 0 {
		return true
	}
	return false
}
