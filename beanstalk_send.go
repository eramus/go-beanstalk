package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eramus/worker"
)

var (
	host    = flag.String("host", "0.0.0.0:11300", "beanstalkd host")
	delay   = flag.Duration("delay", (250 * time.Millisecond), "delay between job requests")
	workers = flag.Int("workers", 1, "numbers of workers to run")
)

const addTube = `example_add`

type addData struct {
	A int `json:"a"`
	B int `json:"b"`
}

func run(shutdown <-chan struct{}, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	c, err := net.Dial("tcp", *host)
	if err != nil {
		log.Println("conn err", err)
		return
	}
	defer c.Close()

	a := &addData{}

	options := worker.GetDefaults()
	options.Conn = c

	for {
		select {
		case <-shutdown:
			return
		default:
		}

		a.A = rand.Intn(10)
		a.B = rand.Intn(10)

		log.Println("Send:", a.A, a.B)

		_, err := worker.Send(addTube, a, false, &options)
		if err != nil {
			log.Println("err:", err)
			return
		}

		<-time.After(*delay)
	}
}

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	var shutdown = make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var (
		finished = make(chan struct{})
		done     = make(chan struct{})
	)

	for i := 0; i < *workers; i++ {
		go run(finished, done)
	}

	<-shutdown

	close(finished)

	for i := 0; i < *workers; i++ {
		<-done
	}
}
