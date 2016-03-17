package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eramus/worker"
)

var (
	host    = flag.String("host", "0.0.0.0:11300", "beanstalkd host")
	reserve = flag.Duration("reserve", time.Second, "delay between job requests")
	workers = flag.Int("workers", 2, "numbers of workers to run")
)

const addTube = `example_add`

type addData struct {
	A int `json:"a"`
	B int `json:"b"`
}

func adder(req *worker.Request) (res worker.Response) {
	a := &addData{}
	err := json.Unmarshal(req.Data, a)
	if err != nil {
		return req.RetryJob(err, 3, nil)
	}

	fmt.Printf("Add: %d + %d = %d\n\n", a.A, a.B, (a.A + a.B))
	return
}

func main() {
	flag.Parse()

	options := worker.GetDefaults()
	options.Host = *host
	options.Reserve = *reserve
	options.Count = *workers

	add := worker.New(addTube, adder, &options)
	add.Run()

	var shutdown = make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Println("Running worker")
	<-shutdown

	f := make(chan struct{})
	add.Shutdown(f)
	<-f
}
