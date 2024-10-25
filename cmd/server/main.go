package main

import (
	"errors"
	"flag"
	"html/template"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/Jeskay/musthave_metrics/internal/metric/routes"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"

	"github.com/Jeskay/musthave_metrics/internal/metric"
)

var conf = config.NewServerConfig()

func main() {
	t, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}
	zapL := zap.Must(zap.NewProduction())
	fs, err := internal.NewFileStorage(conf.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	service := metric.NewMetricService(*conf, zapslog.NewHandler(zapL.Core(), nil), fs, internal.NewMemStorage())

	r := routes.Init(service, t)

	r.Run(conf.Address)
	service.StartSaving()
	defer service.Close()
}

func init() {
	flag.IntVar(&conf.SaveInterval, "i", conf.SaveInterval, "save to storage interval")
	flag.StringVar(&conf.StoragePath, "f", conf.StoragePath, "path to storage file")
	flag.BoolVar(&conf.Restore, "r", conf.Restore, "load values from existing file on start")
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
