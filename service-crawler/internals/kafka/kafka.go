package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/usercommon/crawler/internals"
)

type DataTask struct {
	Html      string `json:"html"`
	Url       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

func SendToKafka(data internals.Data) error {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_ADDR")),
		Topic:                  "web_pages_unprocessed",
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		Compression:            kafka.Snappy,
		BatchSize:              100,
		BatchBytes:             50e6,
		WriteTimeout:           10 * time.Minute,
	}
	defer writer.Close()

	task := DataTask{
		Html:      data.Html,
		Url:       data.Url,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	payload, _ := json.Marshal(task)
	err := writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(data.Url),
			Value: payload,
		},
	)

	return err
}

func HandleSendToKafka(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_ADDR")),
		Topic:                  "web_pages_unprocessed",
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		Compression:            kafka.Snappy,
		BatchSize:              100,
		BatchBytes:             50e6,
		WriteTimeout:           10 * time.Minute,
	}

	// 1. Достаем 100 свежих ID
	var ids []int64
	err := db.Select(&ids, "SELECT id FROM pages WHERE is_sent = FALSE LIMIT 100")
	if err != nil || len(ids) == 0 {
		w.Write([]byte("No new incomes"))
		return
	}

	// 2. Готовим сообщения для Кафки
	var messages []kafka.Message
	for _, id := range ids {
		messages = append(messages, kafka.Message{
			Value: []byte(strconv.FormatInt(id, 10)),
		})
	}

	// 3. Пушим в Кафку
	err = writer.WriteMessages(context.Background(), messages...)
	if err != nil {
		http.Error(w, "Кафка сбоит", 500)
		return
	}

	// 4. Помечаем как отправленные
	db.Exec("UPDATE pages SET is_sent = TRUE WHERE id = ANY($1)", pq.Array(ids))

	w.Write([]byte(fmt.Sprintf("Отправлено в Кафку: %d записей", len(ids))))
}
