package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/usercommon/crawler/internals"
)

func main() {
	workersCount := flag.Uint("w", 10, "Amount of simultaneously running workers. 10 by default.")
	startUrl := flag.String("url", "https://gobyexample.com/structs", "Start crawling from this url.")

	w := internals.Init(uint32(*workersCount))
	w.Run(func(d internals.Data) {
		fmt.Printf("Found some data! url: %v\n", d.Url)
	}, *startUrl)
	defer w.Close()

	for {
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("Ended!")
}
