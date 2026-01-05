package main

import (
	"fmt"
	"time"

	"github.com/usercommon/crawler/internals"
)

func main() {
	w := internals.Init(10)

	w.Run(func(d internals.Data) {
		fmt.Printf("Found some data! url: %v\n", d.Url)
	}, "https://gobyexample.com/structs")
	defer w.Close()

	for {
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("Ended!")
}
