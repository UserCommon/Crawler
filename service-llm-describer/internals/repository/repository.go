package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

func SaveResults(db *sqlx.DB, data AnalysisResult) error {
	// Конвертируем []float32 в формат Postgres vector: "[0.1,0.2,0.3]"

	query := `
		INSERT INTO pages (
			url,
			html,
			structure_analysis,
			content_analysis,
			embedding
		) VALUES (
			:url,
			:html,
			:structure_analysis,
			:content_analysis,
			:embedding
		)
		ON CONFLICT (url) DO UPDATE SET
			structure_analysis = EXCLUDED.structure_analysis,
			content_analysis = EXCLUDED.content_analysis,
			embedding = EXCLUDED.embedding
	`

	// Используем NamedExec для основной структуры + передаем вектор отдельно
	// Но проще использовать NamedExec, добавив вектор в мапу или временную структуру:

	args := map[string]interface{}{
		"url":                data.Url,
		"html":               data.Html,
		"structure_analysis": data.StructureAnalysis,
		"content_analysis":   data.ContentAnalysis,
		"embedding":          pgvector.NewVector(data.Embedding),
	}

	_, err := db.NamedExec(query, args)
	return err
}
