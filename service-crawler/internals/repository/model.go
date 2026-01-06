package repository

import (
	"time"
)

type Page struct {
	ID        int64     `db:"id"`
	URL       string    `db:"url"`
	RawHTML   string    `db:"raw_html"`
	IsSent    bool      `db:"is_sent"`
	CreatedAt time.Time `db:"created_at"`
}
