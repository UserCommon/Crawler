package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func SavePage(db *sqlx.DB, url, html string) error {
	query := `INSERT INTO pages(url, raw_html) VALUES ($1, $2)
			  ON CONFLICT (url) DO UPDATE SET raw_html = EXCLUDED.raw_html, is_sent = FALSE`
	_, err := db.Exec(query, url, html)
	return err
}
