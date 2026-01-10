package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/usercommon/llm-describer/internals/db"
	"github.com/usercommon/llm-describer/internals/repository"
	"github.com/usercommon/llm-describer/internals/worker"
	"github.com/usercommon/llm-describer/proto"
)

// FIXME: Something wrong with first pull of kafka
func main() {
	checkEnvs("GRPC_CRAWLER_HOST", "GRPC_CRAWLER_PORT", "OLLAMA_HOST", "OLLAMA_PORT", "KAFKA_ADDR")

	dbConn := db.InitDB()

	// 1. Setup gRPC connection to Crawler
	grpcUrl := fmt.Sprintf("%s:%s", os.Getenv("GRPC_CRAWLER_HOST"), os.Getenv("GRPC_CRAWLER_PORT"))
	conn, err := grpc.NewClient(grpcUrl,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(20*1024*1024)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("gRPC connection error: %v", err)
	}
	defer conn.Close()
	crawlerClient := proto.NewCrawlerServiceClient(conn)

	// 2. Initialize LLM Processor
	ollamaUrl := fmt.Sprintf("http://%s:%s", os.Getenv("OLLAMA_HOST"), os.Getenv("OLLAMA_PORT"))
	proc, err := worker.NewProcessor(ollamaUrl)
	if err != nil {
		log.Fatalf("LLM initialization error: %v", err)
	}

	// 3. Setup Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{os.Getenv("KAFKA_ADDR")},
		Topic:          "web_pages_unprocessed",
		GroupID:        "llm-worker-1",
		CommitInterval: 1 * time.Second,
	})
	defer reader.Close()

	log.Println("Worker manager started successfully")

	// 4. Initialize Worker Pool
	// Adjust numWorkers based on your RTX 4060 VRAM and performance
	numWorkers, err := strconv.Atoi(os.Getenv("SERVICE_LLM_WORKERS"))
	if err != nil {
		log.Fatalf("Failed to read env var: SERVICE_LLM_WORKERS")
	}
	jobs := make(chan kafka.Message)

	for w := 1; w <= numWorkers; w++ {
		go startWorker(w, jobs, crawlerClient, proc, dbConn)
	}

	// 5. Main loop: Fetch from Kafka and dispatch to workers
	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := reader.ReadMessage(readCtx)

		// TODO: figure out what's wrong.
		_, err = crawlerClient.SendToKafka(context.Background(), &emptypb.Empty{})
		if err != nil {
			log.Printf("shit")
		}
		readCancel()

		if err != nil {
			log.Printf("Kafka read error: %v. Sending refill request to Crawler.", err)
			_, err := crawlerClient.SendToKafka(context.Background(), &emptypb.Empty{})
			if err != nil {
				log.Printf("SendToKafka refill error: %v", err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		// Send task to the pool
		jobs <- msg
	}
}

func startWorker(id int, jobs <-chan kafka.Message, client proto.CrawlerServiceClient, proc *worker.Processor, db *sqlx.DB) {
	var sem = make(chan struct{}, 3)
	log.Printf("Worker [%d] initialized", id)
	for msg := range jobs {
		// Parse Page ID
		pageIDStr := string(msg.Value)
		pageID, err := strconv.ParseInt(pageIDStr, 10, 64)
		if err != nil {
			log.Printf("Worker [%d] Error: Invalid ID in Kafka (%s): %v", id, pageIDStr, err)
			continue
		}

		log.Printf("Worker [%d] processing ID %d", id, pageID)

		workerCtx, workerCancel := context.WithTimeout(context.Background(), 120*time.Second)
		pageData, err := client.GetPage(workerCtx, &proto.PageRequest{Id: pageID})

		if err != nil {
			log.Printf("Worker [%d] gRPC error for ID %d: %v", id, pageID, err)
			workerCancel()
			continue
		}

		pageMd, err := proc.GetMarkdown(pageData.Html)
		if err != nil {
			log.Printf("Worker [%d] gRPC error for ID %d: %v", id, pageID, err)
			workerCancel()
			continue
		}

		// Perform LLM Analysis

		var structRes, contentRes string
		var embeddingsRes []float32
		// Three tasks: Analyze structure; Analyze content; Get embeddings
		g, gCtx := errgroup.WithContext(workerCtx)

		// get structure
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			var err error
			structRes, err = proc.AnalyzeStructure(gCtx, pageData.Html)
			return err
		})

		// get content
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			var err error
			contentRes, err = proc.AnalyzeContent(gCtx, pageMd)
			return err
		})

		// get embeddings
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			var err error
			embeddingsRes, err = proc.GenerateEmbedding(gCtx, pageMd)
			return err
		})

		// Wait until end
		if err := g.Wait(); err != nil {
			log.Printf("Worker [%d] ID %d failed: %v", id, pageID, err)
			workerCancel()
			continue
		}

		// Write to DB
		rec := repository.AnalysisResult{
			Url:               pageData.Url,
			Html:              pageData.Html,
			StructureAnalysis: structRes,
			ContentAnalysis:   contentRes,
			Embedding:         embeddingsRes,
		}
		log.Printf("Worker [%d] finished ID %d. Type: %s | Summary: %s", id, pageID, structRes, contentRes)
		if err := repository.SaveResults(db, rec); err != nil {
			log.Printf("Worker [%d] DB error: %v", id, err)
		} else {
			log.Printf("Worker [%d] ID %d SUCCESS", id, pageID)
		}
		workerCancel()
	}
}

func checkEnvs(keys ...string) {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			log.Fatalf("Environment variable %s is not set", k)
		}
	}
}
