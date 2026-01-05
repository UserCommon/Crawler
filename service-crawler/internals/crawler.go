package internals

import (
	"log"
	"sync"

	"github.com/usercommon/crawler/internals/html_utils"
)

type Crawler struct {
	Workers []*Worker
	// As much tokens as much workers there are
	// For limiting how much workers should crawl web due to
	// socket limits
	DataPool chan Data
	UrlPool  chan string
	Tokens   chan struct{}
}

func Init(workers uint32) *Crawler {
	/// Initialize Crawler with worker amount
	workers_array := make([]*Worker, workers)
	for id, _ := range workers_array {
		workers_array[id] = WorkerInit(uint32(id))
	}

	return &Crawler{
		Workers:  workers_array,
		DataPool: make(chan Data, 1024),
		UrlPool:  make(chan string, 4096),
		Tokens:   make(chan struct{}, workers),
	}
}

func (self *Crawler) Close() {
	close(self.DataPool)
	close(self.UrlPool)
	close(self.Tokens)
}

func (self *Crawler) Run(on_write func(Data), startURL string) error {
	on_error := func(e error) { log.Printf("error: %v", e) }

	taskWg := &sync.WaitGroup{}

	for _, w := range self.Workers {
		go w.runWorker(self.Tokens, self.UrlPool, self.DataPool, on_error, taskWg)
	}

	// Send first url
	taskWg.Add(1)
	self.UrlPool <- startURL

	// No more url
	go func() {
		taskWg.Wait()
		close(self.UrlPool)
		close(self.DataPool)
	}()

	// results
	for data := range self.DataPool {
		on_write(data)
	}

	return nil
}

type Worker struct {
	id uint32
}

func WorkerInit(id uint32) *Worker {
	return &Worker{
		id,
	}
}

func (self *Worker) runWorker(
	tokens chan struct{},
	url_pool chan string,
	data_pool chan<- Data,
	on_error func(error),
	taskWg *sync.WaitGroup, // Передаем счетчик задач
) {
	// Токен ограничивает количество ОДНОВРЕМЕННЫХ запросов в сеть
	for url := range url_pool {
		// Берем токен ПЕРЕД сетевым запросом
		tokens <- struct{}{}

		html, err := html_utils.FetchHtml(url)
		if err != nil {
			on_error(err)
			<-tokens
			taskWg.Done() // Ошибка — это тоже завершение задачи
			continue
		}

		node, err := html_utils.ParseHtml(html)
		links, _ := html_utils.ExtractLinks(node, url)

		<-tokens // Отпускаем сетевой токен

		data_pool <- Data{html, url}

		// Добавляем новые ссылки в очередь
		for _, link := range links {
			taskWg.Add(1)
			// Используем горутину, чтобы не заблокировать воркера,
			// если канал url_pool полон
			go func(l string) {
				url_pool <- l
			}(link)
		}

		// Помечаем, что текущая ссылка полностью обработана
		taskWg.Done()
	}
}

type Data struct {
	Html string `json:"html"`
	Url  string `json:"url"`
}
