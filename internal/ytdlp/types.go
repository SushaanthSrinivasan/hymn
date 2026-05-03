package ytdlp

import "time"

type Result struct {
	VideoID      string
	URL          string
	Title        string
	Artist       string
	Duration     time.Duration
	ThumbnailURL string
	IsLive       bool
}

func (r Result) DurationString() string {
	if r.Duration <= 0 {
		return "--:--"
	}
	total := int(r.Duration.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return formatTime(h, m, s)
	}
	return formatTime(0, m, s)[3:]
}

func formatTime(h, m, s int) string {
	const digits = "0123456789"
	buf := []byte{
		digits[(h/10)%10], digits[h%10], ':',
		digits[(m/10)%10], digits[m%10], ':',
		digits[(s/10)%10], digits[s%10],
	}
	return string(buf)
}
