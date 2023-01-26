package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/e-inwork-com/go-profile-indexing-service/internal/data"
	"github.com/e-inwork-com/go-profile-indexing-service/internal/jsonlog"
	"github.com/e-inwork-com/go-profile-indexing-service/internal/profiles"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

var (
	BuildTime string
	Version   string
)

type Config struct {
	Env string

	Db struct {
		Dsn         string
		MaxOpenConn int
		MaxIdleConn int
		MaxIdleTime string
	}

	GRPCPort    string
	SolrURL     string
	SolrProfile string
}

type Application struct {
	Config  Config
	Logger  *jsonlog.Logger
	Models  data.Models
	Indexes data.Indexes
}

type ProfileServer struct {
	profiles.UnimplementedProfileServiceServer
	Indexes data.Indexes
	Models  data.Models
}

func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Db.MaxOpenConn)
	db.SetMaxIdleConns(cfg.Db.MaxIdleConn)

	duration, err := time.ParseDuration(cfg.Db.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (t *ProfileServer) WriteProfile(ctx context.Context, req *profiles.ProfileRequest) (*profiles.ProfileResponse, error) {
	input := req.GetProfileEntry()

	id, err := uuid.Parse(input.Id)
	if err != nil {
		return nil, err
	}

	profile, err := t.Models.Profiles.Get(id)
	if err != nil {
		res := &profiles.ProfileResponse{Result: "Failed"}
		return res, err
	}

	if profile.IsIndexed {
		return &profiles.ProfileResponse{Result: "Indexed"}, nil
	}

	resp, err := t.Indexes.Profiles.Update(profile)
	if err != nil || resp.StatusCode != http.StatusOK {
		return &profiles.ProfileResponse{Result: "Failed"}, err
	}

	if !profile.IsDeleted {
		err = t.Models.Profiles.IsIndexedTrue(profile)
		if err != nil {
			return &profiles.ProfileResponse{Result: "Failed"}, err
		}
	} else {
		err = t.Models.Profiles.Delete(profile)
		if err != nil {
			return &profiles.ProfileResponse{Result: "Failed"}, err
		}
	}

	return &profiles.ProfileResponse{Result: "Indexed"}, nil
}

func (app *Application) GRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", app.Config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()

	profiles.RegisterProfileServiceServer(s, &ProfileServer{
		Indexes: app.Indexes,
		Models:  app.Models})

	log.Printf("gRPC Server started on port: %v", app.Config.GRPCPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}
