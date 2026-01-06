package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/usercommon/llm-describer/internals/worker"
	"github.com/usercommon/llm-describer/proto"
)

func main() {
	checkEnvs("GRPC_CRAWLER_HOST", "GRPC_CRAWLER_PORT", "OLLAMA_HOST", "OLLAMA_PORT", "KAFKA_BROKERS")

	// 1. Connect to Crawler Service
	grpcUrl := fmt.Sprintf("%s:%s", os.Getenv("GRPC_CRAWLER_HOST"), os.Getenv("GRPC_CRAWLER_PORT"))
	conn, err := grpc.NewClient(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ gRPC connection error: %v", err)
	}
	defer conn.Close()
	crawlerClient := proto.NewCrawlerServiceClient(conn)

	// 2. Initialize LLM
	ollamaUrl := fmt.Sprintf("http://%s:%s", os.Getenv("OLLAMA_HOST"), os.Getenv("OLLAMA_PORT"))
	proc, err := worker.NewProcessor(ollamaUrl)
	if err != nil {
		log.Fatalf("❌ LLM init error: %v", err)
	}

	// 3. Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_ADDR")},
		Topic:   "web_pages_unprocessed",
		GroupID: fmt.Sprintf("llm-worker-%d", time.Now().Unix()),
	})
	defer reader.Close()
	fmt.Println(os.Getenv("KAFKA_ADDR"))

	log.Println("Worker has started!")

	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 60*time.Second)
		msg, err := reader.ReadMessage(readCtx)
		readCancel()

		if err != nil {
			log.Printf("DEBUG KAFKA: ошибка чтения: %v\n", err)
			log.Println("Empty queue, sending request to refill")
			_, err := crawlerClient.SendToKafka(context.Background(), &emptypb.Empty{})
			if err != nil {
				log.Printf("Error SendToKafka: %v\n", err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		// get Id
		pageIDStr := string(msg.Value)
		pageID, err := strconv.ParseInt(pageIDStr, 10, 64)
		if err != nil {
			log.Printf("❌ Кривой ID в Кафке (%s): %v\n", pageIDStr, err)
			continue
		}

		log.Printf("📦 Взял в работу ID %d. Запрашиваю HTML...\n", pageID)

		grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 30*time.Second)
		pageData, err := crawlerClient.GetPage(grpcCtx, &proto.PageRequest{Id: pageID})
		grpcCancel()

		if err != nil {
			log.Printf("❌ Ошибка gRPC GetPage (ID %d): %v\n", pageID, err)
			continue
		}

		// --- АНАЛИЗ ЧЕРЕЗ LLM (Таймаут 60 сек на оба запроса) ---
		llmCtx, llmCancel := context.WithTimeout(context.Background(), 60*time.Second)

		log.Printf("🧠 [ID %d] Отправляю на 4060 (анализ структуры)...\n", pageID)
		structRes, err := proc.AnalyzeStructure(llmCtx, pageData.Html)
		if err != nil {
			log.Printf("❌ [ID %d] Ошибка структуры: %v", pageID, err)
		}

		log.Printf("🧠 [ID %d] Отправляю на 4060 (анализ смысла)...\n", pageID)
		contentRes, err := proc.AnalyzeContent(llmCtx, pageData.Html)
		if err != nil {
			log.Printf("❌ [ID %d] Ошибка контента: %v\n", pageID, err)
		}
		llmCancel()

		log.Printf("✅ ГОТОВО [ID %d] | Type:\n %s\n | Summary: \n%s\n", pageID, structRes, contentRes)
	}
}

func checkEnvs(keys ...string) {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			log.Fatalf("Environment variable %s is not set", k)
		}
	}
}
