package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
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

var conf *config.ServerConfig

var buildVersion string
var buildDate string
var buildCommit string

func main() {

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

	service := initHelper(conf, zapL)

	r := routes.Init(conf, service, t)
	go func() {
		if err := r.Run(conf.Address); err != nil && err != http.ErrServerClosed {
			zapL.Fatal("server stopped", zap.Error(err))
		}
	}()
	service.StartSaving()

	shutdownHelper(context.Background(), service, zapL)
}

func initHelper(conf *config.ServerConfig, logger *zap.Logger) *metric.MetricService {
	var storage internal.Repositories

	fs, err := db.NewFileStorage(conf.StoragePath)
	if err != nil {
		logger.Fatal("failed to init file storage", zap.Error(err))
	}

	if conf.DBConnection == "" {
		storage = db.NewMemStorage()
	} else {
		database, err := sql.Open("pgx", conf.DBConnection)
		if err != nil {
			logger.Fatal("failed to connect to database", zap.Error(err))
		}
		if storage, err = db.NewPostgresStorage(database, zapslog.NewHandler(logger.Core(), nil)); err != nil {
			logger.Fatal("failed to init database", zap.Error(err))
		}
	}
	return metric.NewMetricService(*conf, zapslog.NewHandler(logger.Core(), nil), fs, storage)
}

func shutdownHelper(ctx context.Context, service *metric.MetricService, logger *zap.Logger) {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-sig
	logger.Info("initiating server shutdown...")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	service.Close()
	<-ctx.Done()
	logger.Info("server shutting down")
}

func init() {
	confParam := loadParams()
	if err := env.Parse(confParam); err != nil {
		log.Fatal(err)
	}
	if confParam.Config != "" {
		b, err := os.ReadFile(confParam.Config)
		if err != nil {
			return
		}
		var confJSON = config.NewServerConfig()
		err = json.Unmarshal(b, confJSON)
		if err != nil {
			return
		}
		confParam.Merge(confJSON)
	}
	conf = confParam
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

func loadParams() *config.ServerConfig {
	var paramCfg = config.NewServerConfig()

	flag.IntVar(&paramCfg.SaveInterval, "i", paramCfg.SaveInterval, "save to storage interval")
	flag.StringVar(&paramCfg.DBConnection, "d", "", "database connection string")
	flag.StringVar(&paramCfg.HashKey, "k", "", "secret hash key")
	flag.StringVar(&paramCfg.TLSPrivate, "crypto-key", "", "path to cryptographic key file")
	flag.StringVar(&paramCfg.Config, "config", "", "path to configuration file")
	flag.Func("f", "storage file location", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		if !util.IsValidPath(s) {
			return errors.New("invalid path format")
		}
		paramCfg.StoragePath = s
		return nil
	})
	flag.BoolVar(&paramCfg.Restore, "r", paramCfg.Restore, "load values from existing file on start")
	flag.Func("a", "server address", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		ok, err := regexp.Match(`[^\:]*:[0-9]{4}`, []byte(s))
		if !ok {
			return errors.New("invalid address format")
		}
		paramCfg.Address = s
		return err
	})

	flag.Parse()
	return paramCfg
}
