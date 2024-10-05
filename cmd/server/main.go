package main

import (
	"errors"
	"flag"
	"log"
	"regexp"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/metric/routes"
	"github.com/caarlos0/env"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

var conf = config.NewServerConfig()

func main() {
	service := metric.NewMetricService()
	r := routes.Init(service)

	r.Run(conf.Address)
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
		conf.Address = s
		return err
	})

	flag.Parse()
	if err := env.Parse(conf); err != nil {
		log.Fatal(err)
	}
}
