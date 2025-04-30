package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/agent"
	"github.com/Jeskay/musthave_metrics/internal/util"
	"github.com/caarlos0/env"
)

var conf = config.NewAgentConfig()

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	sig := make(chan os.Signal, 1)
	var endMonitor, endSender chan<- struct{}

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s \nBuild date: %s \nBuild commit: %s \n", buildVersion, buildDate, buildCommit)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	client := &http.Client{
		Timeout: 6 * time.Second,
	}
	logger := slog.NewTextHandler(os.Stdout, nil)
	svc := agent.NewAgentService(client, conf, logger)
	err := util.TryRun(func() error {
		return svc.CheckAPIAvailability()
	}, util.IsConnectionRefused)

	if err != nil {
		slog.Error(err.Error())
	}
	endMonitor = svc.StartMonitoring(conf.GetReportInterval())
	endSender = svc.StartSending(conf.GetPollInterval())
	<-sig
	endMonitor <- struct{}{}
	endSender <- struct{}{}
}

func init() {
	flag.IntVar(&conf.ReportInterval, "r", 10, "report frequency in seconds")
	flag.IntVar(&conf.RateLimit, "l", 1, "amount of concurrent requests to server")
	flag.StringVar(&conf.HashKey, "k", "", "secret hash key")
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
