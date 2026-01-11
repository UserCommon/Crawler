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

func SearchPages(db *sqlx.DB, queryVector []float32, limit int) ([]AnalysisResponse, error) {
	var results []AnalysisResponse

	// 1. Добавляем дистанцию в SELECT
	query := `
        SELECT
            url,
            content_analysis,
            (embedding <=> :v) as distance
        FROM pages
        ORDER BY distance ASC
        LIMIT :limit`

	args := map[string]interface{}{
		"v":     pgvector.NewVector(queryVector),
		"limit": limit,
	}

	rows, err := db.NamedQuery(query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r AnalysisResponse
		// 2. Используем StructScan вместо обычного Scan
		// Он автоматически найдет поля по тегам db:"url", db:"content_analysis" и db:"distance"
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// Search page
// func SearchPages(db *sqlx.DB, queryVector []float32, limit int) ([]AnalysisResponse, error) {
// 	var results []AnalysisResponse
// 	query := `
//         SELECT
//         	url,
//          	content_analysis,
// 			(embedding <=> :v) as distance
//         FROM pages
//         ORDER BY embedding <=> :v
//         LIMIT :limit`

// 	args := map[string]interface{}{
// 		"v":     pgvector.NewVector(queryVector),
// 		"limit": limit,
// 	}

// 	// Используем хелпер для именованных параметров
// 	rows, err := db.NamedQuery(query, args)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var r AnalysisResponse
// 		if err := rows.Scan(&r.Url, &r.ContentAnalysis); err != nil {
// 			return nil, err
// 		}
// 		results = append(results, r)
// 	}
// 	log.Printf("F: %v\n", results)
// 	return results, nil

// }
