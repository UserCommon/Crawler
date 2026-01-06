// package worker

// import (
// 	md "github.com/JohannesKaufmann/html-to-markdown"
// 	"github.com/tmc/langchaingo/llms/ollama"
// 	"github.com/usercommon/llm-describer/proto" // твой сгенерированный gRPC код
// )

// type Processor struct {
// 	llm        *ollama.LLM
// 	grpcClient proto.CrawlerServiceClient
// 	converter  *md.Converter
// }

// func NewProcessor(ollamaURL string, grpcClient proto.CrawlerServiceClient) (*Processor, error) {
// 	// Подключаемся к Ollama через LangChainGo
// 	llm, err := ollama.New(ollama.WithModel("nemotron-mini"), ollama.WithServerURL(ollamaURL))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Processor{
// 		llm:        llm,
// 		grpcClient: grpcClient,
// 		converter:  md.NewConverter("", true, nil),
// 	}, nil
// }
