package urls

import "time"

type URL struct {
	ShortUrl  string `gorm:"primaryKey;uniqueIndex:idx_short"`
	LongUrl   string
	UsedCount int `gorm:"default:0"`
	CreatedAt time.Time
}
