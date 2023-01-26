package grpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/e-inwork-com/go-profile-indexing-service/internal/data"
	"github.com/e-inwork-com/go-profile-indexing-service/internal/jsonlog"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	// Configuration
	var cfg Config

	cfg.Db.Dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	cfg.Db.MaxOpenConn = 25
	cfg.Db.MaxIdleConn = 25
	cfg.Db.MaxIdleTime = "15m"
	cfg.GRPCPort = "5002"
	cfg.SolrURL = "http://localhost:8983"
	cfg.SolrProfile = "profiles"

	// Logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Database
	db, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	// Read SQL file
	script, err := os.ReadFile("./test/sql/delete_all.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Delete Records
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Delete Indexing
	res, err := http.Post(
		"http://localhost:8983/solr/"+cfg.SolrProfile+"/update?commit=true",
		"application/json",
		bytes.NewReader([]byte("{'delete': {'query': '*:*'}}")))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatal(res)
	}

	// Application
	app := Application{
		Config:  cfg,
		Logger:  logger,
		Models:  data.InitModels(db),
		Indexes: data.InitIndexes(cfg.SolrURL, cfg.SolrProfile),
	}

	// Run gRPC
	go app.GRPCListen()

	// Initial
	email := "jon@doe.com"
	password := "pa55word"

	// Initial
	client := &http.Client{}
	var userResponse map[string]interface{}

	t.Run("Register User", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email": "%v", "password": "%v", "first_name": "Jon", "last_name": "Doe"}`,
			email,
			password)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8000/service/users",
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &userResponse)
		assert.Nil(t, err)
	})

	// Initial
	type authType struct {
		Token string `json:"token"`
	}
	var authentication authType

	t.Run("Login User", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email": "%v", "password": "%v"}`,
			email,
			password)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8000/service/users/authentication",
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &authentication)
		assert.Nil(t, err)
		assert.NotNil(t, authentication.Token)
	})

	// Initial
	var profileResponse map[string]interface{}

	t.Run("Create Profile", func(t *testing.T) {
		tBody, tContentType := app.testFormProfile(t)
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8000/service/profiles",
			tBody)
		req.Header.Add("Content-Type", tContentType)

		bearer := fmt.Sprintf("Bearer %v", authentication.Token)
		req.Header.Set("Authorization", bearer)

		res, err := client.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &profileResponse)
		assert.Nil(t, err)
	})

	t.Run("Get Profile on Solr", func(t *testing.T) {
		// We use gRPC to update a Profile on the Solr,
		// so it needs to wait a couple seconds until the updating is done
		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest(
			"GET",
			"http://localhost:8983/api/collections/"+cfg.SolrProfile+"/select?q=*:*",
			nil)

		res, err := client.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(body), &result)
		assert.Nil(t, err)

		response := result["response"].(map[string]interface{})
		assert.NotNil(t, response)
		assert.Equal(t, response["numFound"], float64(1))

		docs := response["docs"].([]interface{})
		assert.NotNil(t, docs)

		doc := docs[0].(map[string]interface{})
		assert.NotNil(t, doc)
		assert.NotNil(t, doc["profile_name"])
	})
}
