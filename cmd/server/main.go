package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/pkg/profile"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"

	"github.com/Jeskay/musthave_metrics/config"
	"github.com/Jeskay/musthave_metrics/internal"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/Jeskay/musthave_metrics/internal/metric/db"
	"github.com/Jeskay/musthave_metrics/internal/metric/routes"
	"github.com/Jeskay/musthave_metrics/internal/util"
)

var conf = config.NewServerConfig()

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	var storage internal.Repositories

	prof := profile.Start(profile.MemProfile)
	time.AfterFunc(time.Second*30, prof.Stop)

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

	zapL := zap.Must(zap.NewProduction())
	t, err := loadTemplate()
	if err != nil {
		zapL.Fatal("failed to load templates", zap.Error(err))
	}

	fs, err := db.NewFileStorage(conf.StoragePath)
	if err != nil {
		zapL.Fatal("failed to init file storage", zap.Error(err))
	}

	if conf.DBConnection == "" {
		storage = db.NewMemStorage()
	} else {
		database, err := sql.Open("pgx", conf.DBConnection)
		if err != nil {
			zapL.Fatal("failed to connect to database", zap.Error(err))
		}
		if storage, err = db.NewPostgresStorage(database, zapslog.NewHandler(zapL.Core(), nil)); err != nil {
			zapL.Fatal("failed to init database", zap.Error(err))
		}
	}

	service := metric.NewMetricService(*conf, zapslog.NewHandler(zapL.Core(), nil), fs, storage)

	r := routes.Init(conf.HashKey, service, t)

	r.Run(conf.Address)
	service.StartSaving()
	defer service.Close()
}

func init() {
	flag.IntVar(&conf.SaveInterval, "i", conf.SaveInterval, "save to storage interval")
	flag.StringVar(&conf.DBConnection, "d", "", "database connection string")
	flag.StringVar(&conf.HashKey, "k", "", "secret hash key")
	flag.Func("f", "storage file location", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		if !util.IsValidPath(s) {
			return errors.New("invalid path format")
		}
		conf.StoragePath = s
		return nil
	})
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
