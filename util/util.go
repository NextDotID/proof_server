package util

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/nextdotid/proof_server/util/base1024"

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
