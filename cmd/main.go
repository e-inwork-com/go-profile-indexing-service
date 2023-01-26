package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/e-inwork-com/go-profile-indexing-service/grpc"
	"github.com/e-inwork-com/go-profile-indexing-service/internal/data"
	"github.com/e-inwork-com/go-profile-indexing-service/internal/jsonlog"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	// Load .env if available
	err := godotenv.Load()
	if err != nil {
		log.Println("Enviroment file .env is not found!")
	}

	// Set Configuration
	var cfg grpc.Config

	// Read environment  from a command line and OS
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", os.Getenv("DBDSN"), "Database DSN")
	flag.IntVar(&cfg.Db.MaxOpenConn, "db-max-open-conn", 25, "Database max open connections")
	flag.IntVar(&cfg.Db.MaxIdleConn, "db-max-idle-conn", 25, "Database max idle connections")
	flag.StringVar(&cfg.Db.MaxIdleTime, "db-max-idle-time", "15m", "Database max connection idle time")
	flag.StringVar(&cfg.GRPCPort, "grpc-port", os.Getenv("GRPCPORT"), "gRPC server port")
	flag.StringVar(&cfg.SolrURL, "solr-url", os.Getenv("SOLRURL"), "Solr URL")
	flag.StringVar(&cfg.SolrProfile, "solr-profile", os.Getenv("SOLRPROFILE"), "Solr Profile Path")
	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	// Show version on the terminal
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", grpc.Version)
		fmt.Printf("Build time:\t%s\n", grpc.BuildTime)
		os.Exit(0)
	}

	// Set logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Set Database
	db, err := grpc.OpenDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Log a status of the database
	logger.PrintInfo("database connection pool established", nil)

	// Publish variables
	expvar.NewString("version").Set(grpc.Version)
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Set the application
	app := &grpc.Application{
		Config:  cfg,
		Logger:  logger,
		Models:  data.InitModels(db),
		Indexes: data.InitIndexes(cfg.SolrURL, cfg.SolrProfile),
	}

	// Run gRPC Server
	app.GRPCListen()
}
