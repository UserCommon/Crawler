package worker

import (
	"context"
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Processor struct {
	llm       *ollama.LLM
	converter *md.Converter
}

func NewProcessor(ollamaURL string) (*Processor, error) {
	// Инициализируем Ollama
	l, err := ollama.New(ollama.WithModel("nemotron-mini"), ollama.WithServerURL(ollamaURL))
	if err != nil {
		return nil, fmt.Errorf("failed to init ollama: %w", err)
	}

	return &Processor{
		llm:       l,
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

	prompt := fmt.Sprintf("Analyze HTML structure. Return ONLY JSON { \"type\": \"string\" } (values: article, shop, landing, main, other). HTML: %s", html)

	res, err := p.llm.Call(ctx, prompt, llms.WithJSONMode())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res), nil
}

// AnalyzeContent — Запрос №2 (Суть контента)
func (p *Processor) AnalyzeContent(ctx context.Context, html string) (string, error) {

	markdown, err := p.converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("md conversion error: %w", err)
	}

	// Тоже ограничиваем размер
	if len(markdown) > 10000 {
		markdown = markdown[:10000]
	}

	prompt := fmt.Sprintf("Summarize this content in 2 sentences. Return ONLY JSON { \"summary\": \"string\" }. Content: %s", markdown)

	res, err := p.llm.Call(ctx, prompt, llms.WithJSONMode())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res), nil
}
