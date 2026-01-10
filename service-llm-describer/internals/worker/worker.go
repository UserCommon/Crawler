package worker

import (
	"context"
	"fmt"
	"os"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Processor struct {
	llm       *ollama.LLM
	embedder  *ollama.LLM
	converter *md.Converter
}

func NewProcessor(ollamaURL string) (*Processor, error) {
	// Инициализируем Ollama
	llm, err := ollama.New(
		ollama.WithModel(os.Getenv("OLLAMA_MODEL")),
		ollama.WithServerURL(ollamaURL),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init ollama: %w", err)
	}

	embedder, err := ollama.New(
		ollama.WithModel(os.Getenv("OLLAMA_EMBEDDING")),
		ollama.WithServerURL(ollamaURL),
	)

	return &Processor{
		llm:       llm,
		embedder:  embedder,
		converter: md.NewConverter("", true, nil),
	}, nil
}

// AnalyzeStructure — Запрос №1 (Тип страницы)
func (p *Processor) AnalyzeStructure(ctx context.Context, html string) (string, error) {
	// Nemotron-mini не переварит огромный HTML, берем первые 15к символов
	limit := 15000
	if len(html) > limit {
		html = html[:limit]
	}

	systemPrompt := `You are a web classifier. Analyze the provided content.
	Return ONLY a JSON object with these keys:
	- "type": (article, shop, landing, main, or other)
	- "people_mentioned": [(array of mentioned people)]
	- "has_video": boolean of is there video on website or not
	- "mood": Choose the dominant tone from this list:
	    * "humorous": contains jokes, irony, or lighthearted content.
	    * "aggressive": toxic, angry, or confrontational language.
	    * "neutral": formal, scientific, or purely informational.
	    * "clickbait": sensationalist, over-emotional, or provocative.
	    * "helpful": instructional, supportive, or educational.

	Rules:
	- Output MUST be valid JSON.
	- No conversational filler.
	- Do not mention HTML tags.`
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, "Page Content:\n"+html),
	}
	resp, err := p.llm.GenerateContent(ctx, content, llms.WithJSONMode())

	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}

// AnalyzeContent - semantic analysis based on markdown
func (p *Processor) AnalyzeContent(ctx context.Context, markdown string) (string, error) {

	// Тоже ограничиваем размер
	if len(markdown) > 10000 {
		markdown = markdown[:10000]
	}

	systemPrompt := `You are a web structure analyzer.
    Analyze the provided content and return ONLY a valid JSON object.
    Required JSON keys:
    - "type": Choose one from [article, shop, landing, main, other]
    - "summary": A concise 2-sentence summary of the main topic.
    Do not include any conversational text or markdown code blocks.`

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, "Content to analyze: \n"+markdown),
	}

	res, err := p.llm.GenerateContent(ctx, content, llms.WithJSONMode())
	if err != nil {
		return "", err
	}
	return res.Choices[0].Content, nil
}

func (p *Processor) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if len(text) > 8000 {
		text = text[:8000]
	}

	flts, err := p.embedder.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(flts) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return flts[0], nil
}

func (p *Processor) GetMarkdown(html string) (string, error) {
	return p.converter.ConvertString(html)
}
