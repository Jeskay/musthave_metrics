package main

import (
	"errors"
	"flag"
	"html/template"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal/metric/routes"
	"github.com/caarlos0/env"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

var conf = config.NewServerConfig()

func main() {
	t, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.NewTextHandler(os.Stdout, nil)
	service := metric.NewMetricService(logger)
	r := routes.Init(service, t)

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

func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		h, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
