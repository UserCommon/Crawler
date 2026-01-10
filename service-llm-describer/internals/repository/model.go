package repository

type AnalysisResult struct {
	Url               string    `db:"url"`
	Html              string    `db:"html"`
	StructureAnalysis string    `db:"structure_analysis"`
	ContentAnalysis   string    `db:"content_analysis"`
	Embedding         []float32 `db:"-"`
}
