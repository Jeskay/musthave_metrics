package main

import (
	"errors"
	"flag"
	"regexp"

	"github.com/Jeskay/musthave_metrics/internal/metric/routes"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

var address string = ":8080"

func main() {
	service := metric.NewMetricService()
	r := routes.Init(service)

	r.Run(address)
}

func init() {
	flag.Func("a", "server address", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		ok, err := regexp.Match(`[^\:]*:[0-9]{4}`, []byte(s))
		if !ok {
			return errors.New("invalid address format")
		}
		address = s
		return err
	})
	flag.Parse()
}
