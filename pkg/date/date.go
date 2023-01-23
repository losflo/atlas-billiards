package date

import "time"

const (
	solomonFormat = "01/02/2006 03:04:05 PM"
)

func ToSolomonDateFormat(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(solomonFormat)
} // ./toSolomonDateFormat
