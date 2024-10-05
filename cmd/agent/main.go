package main

import (
	"errors"
	"flag"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/Jeskay/musthave_metrics/internal/agent"
)

var (
	address        string = "localhost:8080"
	reportInterval int    = 10
	pollInterval   int    = 2
)

func main() {
	sig := make(chan os.Signal, 1)
	var endMonitor, endSender chan<- bool

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	svc := agent.NewAgentService(address)
	endMonitor = svc.StartMonitoring(time.Second * time.Duration(reportInterval))
	endSender = svc.StartSending(time.Second * time.Duration(pollInterval))
	<-sig
	endMonitor <- true
	endSender <- true
	os.Exit(1)
}

func init() {
	flag.IntVar(&reportInterval, "r", 10, "report frequency in seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll frequency in seconds")
	flag.Func("a", "server address", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		ok, err := regexp.Match(`[^\:]+:[0-9]{4}`, []byte(s))
		if !ok {
			return errors.New("invalid address format")
		}
		address = s
		return err
	})
	flag.Parse()
}
