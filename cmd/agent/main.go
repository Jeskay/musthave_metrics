package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jeskay/musthave_metrics/internal/agent"
)

func main() {
	sig := make(chan os.Signal, 1)
	var endMonitor, endSender chan<- bool

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	svc := agent.NewAgentService("localhost", ":8080")
	endMonitor = svc.StartMonitoring(time.Second * 2)
	endSender = svc.StartSending(time.Second * 4)
	<-sig
	endMonitor <- true
	endSender <- true
	os.Exit(1)
}
