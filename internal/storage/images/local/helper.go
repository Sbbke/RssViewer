package images

import "strings"

type BriefingMeta struct {
	Hash        string  `json:"hash"`
	Timeframe   string  `json:"timeframe"`   // "weekly", "monthly", "custom"
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
	IDs         map[int64][]int64 `json:"ids"`
	NumSlides   int     `json:"numSlides"`
	CreatedAt   string  `json:"createdAt"`
}

func validHash(hash string) bool {
	if hash == "" {
		return false
	}
	for _, c := range hash {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
