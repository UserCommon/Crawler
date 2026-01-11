package repository

type AnalysisResult struct {
	Url               string    `db:"url"`
	Html              string    `db:"html"`
	StructureAnalysis string    `db:"structure_analysis"`
	ContentAnalysis   string    `db:"content_analysis"`
	Embedding         []float32 `db:"-"`
}

type AnalysisResponse struct {
	Url               string  `db:"url"`
	Html              string  `db:"html"`
	StructureAnalysis string  `db:"structure_analysis"`
	ContentAnalysis   string  `db:"content_analysis"`
	Distance          float64 `db:"distance"`
}

type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}
