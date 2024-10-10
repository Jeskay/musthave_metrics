package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/agent"
	"github.com/caarlos0/env"
)

var conf = config.NewAgentConfig()

func main() {
	sig := make(chan os.Signal, 1)
	var endMonitor, endSender chan<- struct{}

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	svc := agent.NewAgentService(conf.Address)
	endMonitor = svc.StartMonitoring(conf.GetReportInterval())
	endSender = svc.StartSending(conf.GetPollInterval())
	<-sig
	endMonitor <- struct{}{}
	endSender <- struct{}{}
	os.Exit(1)
}

func init() {
	flag.IntVar(&conf.ReportInterval, "r", 10, "report frequency in seconds")
	flag.IntVar(&conf.PollInterval, "p", 2, "poll frequency in seconds")
	flag.Func("a", "server address", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		ok, err := regexp.Match(`[^\:]+:[0-9]{4}`, []byte(s))
		if !ok {
			return errors.New("invalid address format")
		}
		conf.Address = s
		return err
	})
	flag.Parse()

	if err := env.Parse(conf); err != nil {
		log.Fatal(err)
	}
}
