package util

import (
	"strconv"
	"time"

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
