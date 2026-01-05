package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/usercommon/crawler/internals"
	"github.com/usercommon/crawler/internals/kafka"
)

func main() {
	workersCount := flag.Uint("w", 10, "Amount of simultaneously running workers. 10 by default.")
	startUrl := flag.String("url", "https://gobyexample.com/structs", "Start crawling from this url.")

	w := internals.Init(uint32(*workersCount))
	w.Run(func(d internals.Data) {
		err := kafka.SendToKafka(d)
		if err != nil {
			fmt.Printf("Failed to send task to kafka! %v\n", err)
		}
	}, *startUrl)
	defer w.Close()

	for {
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("Ended!")
}
