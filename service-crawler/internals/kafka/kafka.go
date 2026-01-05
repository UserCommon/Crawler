package kafka

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/usercommon/crawler/internals"
)

func SendToKafka(data internals.Data) error {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_ADDR")),
		Topic:                  "web_pages_unprocessed",
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		Compression:            kafka.Snappy,
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

type DataTask struct {
	Html      string `json:"html"`
	Url       string `json:"url"`
	CreatedAt string `json:"created_at"`
}
