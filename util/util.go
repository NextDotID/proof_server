package util

import (
	"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/nextdotid/proof_server/util/base1024"
	"github.com/sirupsen/logrus"

	"golang.org/x/xerrors"
)

func TimeToTimestampString(now time.Time) string {
	return strconv.FormatInt(now.Unix(), 10)
}

func TimestampStringToTime(now string) (time.Time, error) {
	ts, err := strconv.ParseInt(now, 10, 64)
	if err != nil {
		return time.Time{}, xerrors.Errorf("%w", err)
	}

	return time.Unix(ts, int64(0)), nil
}

func DecodeString(s string) ([]byte, error) {
	sigBytes, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return sigBytes, nil
	}
	return base1024.DecodeString(s)
}

// GetE gets current system's environment variable as `string`.
// If `defaultValue` is empty, this environment key must be exist, or it will panic.
func GetE(envKey, defaultValue string) string {
	result := os.Getenv(envKey)
	if len(result) == 0 {
		if len(defaultValue) > 0 {
			return defaultValue
		} else {
			logrus.Fatalf("ENV %s must be given! Abort.", envKey)
			return ""
		}

	}
	return result
}
