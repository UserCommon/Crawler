package kafka

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type DataTask struct {
	Html      string `json:"html"`
	Url       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

func SendToKafka(ctx context.Context, db *sqlx.DB) (int, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_ADDR")),
		Topic:                  "web_pages_unprocessed",
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		Compression:            kafka.Snappy,
		BatchSize:              100,
		BatchBytes:             50e6,
		WriteTimeout:           10 * time.Second, // 10 минут было слишком много для синхронного gRPC
	}
	defer writer.Close()

	// 1. Достаем 100 свежих ID
	var ids []int64
	err := db.SelectContext(ctx, &ids, "SELECT id FROM pages WHERE is_sent = FALSE LIMIT 100")
	if err != nil {
		return 0, fmt.Errorf("db select error: %w", err)
	}

	if len(ids) == 0 {
		return 0, nil // Просто нечего отправлять
	}

	// 2. Готовим сообщения
	messages := make([]kafka.Message, len(ids))
	for i, id := range ids {
		messages[i] = kafka.Message{
			Value: []byte(strconv.FormatInt(id, 10)),
		}
	}

	// 3. Пушим в Кафку
	err = writer.WriteMessages(ctx, messages...)
	if err != nil {
		return 0, fmt.Errorf("kafka write error: %w", err)
	}

	// 4. Помечаем как отправленные
	_, err = db.ExecContext(ctx, "UPDATE pages SET is_sent = TRUE WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return len(ids), fmt.Errorf("db update error: %w", err)
	}

	return len(ids), nil
}
