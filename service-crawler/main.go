package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/usercommon/crawler/internals"
	"github.com/usercommon/crawler/internals/db"
	"github.com/usercommon/crawler/internals/kafka"
	"github.com/usercommon/crawler/internals/repository"
	"github.com/usercommon/crawler/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	workersCount := flag.Uint("w", 10, "Amount of simultaneously running workers. 10 by default.")
	startUrl := flag.String("url", "https://gobyexample.com/structs", "Start crawling from this url.")

	dbConn := db.InitDB()
	w := internals.Init(uint32(*workersCount))
	defer w.Close()
	go func() {
		http.HandleFunc("/send_to_kafka", func(w http.ResponseWriter, r *http.Request) {
			kafka.HandleSendToKafka(w, r, dbConn)
			// Тот код, который берет ID из БД, шлет в Кафку
			// и делает UPDATE pages SET is_sent = true
		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	go startGRPC(dbConn)
	err := w.Run(func(d internals.Data) {
		for {
			var count int
			// Считаем сколько еще не отправлено в Кафку
			err := dbConn.Get(&count, "SELECT COUNT(*) FROM pages WHERE is_sent = FALSE")
			if err != nil {
				log.Printf("DB error: %v", err)
				break
			}

			// limit
			if count < 100 {
				break
			}

			// Если в базе завал, просто спим прямо здесь.
			// DataPool забьется, и воркеры встанут.
			log.Printf("Backpressure: в базе уже %d записей. Ждем...", count)
			time.Sleep(10 * time.Second)
		}

		err := repository.SavePage(dbConn, d.Url, d.Html)
		if err != nil {
			fmt.Printf("Failed to save to DB: %v\n", err)
		}
	}, *startUrl)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("Ended!")
}

type crawlerServer struct {
	proto.UnimplementedCrawlerServiceServer
	db *sqlx.DB
}

func (s *crawlerServer) GetPage(ctx context.Context, req *proto.PageRequest) (*proto.PageResponse, error) {
	var page repository.Page
	err := s.db.Get(&page, `SELECT id, url, raw_html FROM pages WHERE id = $1`, req.Id)
	if err != nil {
		return nil, err
	}

	return &proto.PageResponse{
		Id:   page.ID,
		Url:  page.URL,
		Html: page.RawHTML,
	}, nil
}

func startGRPC(db *sqlx.DB) {
	addr := fmt.Sprintf("%s:%s", os.Getenv("GRPC_CRAWLER_HOST"), os.Getenv("GRPC_CRAWLER_PORT"))
	lis, _ := net.Listen("tcp", addr)
	s := grpc.NewServer()
	reflection.Register(s)
	proto.RegisterCrawlerServiceServer(s, &crawlerServer{db: db})
	s.Serve(lis)
}
