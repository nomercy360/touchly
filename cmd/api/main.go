package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	_ "touchly/docs"
	admSvc "touchly/internal/api"
	"touchly/internal/db"
	"touchly/internal/services"
	"touchly/internal/transport"
)

// Config represents the configuration structure
type Config struct {
	DbConnString string       `yaml:"db_conn_string"`
	Server       ServerConfig `yaml:"server"`
	JWTSecret    string       `yaml:"jwt_secret"`
	ResendApiKey string       `yaml:"resend_api_key"`
}

// ServerConfig represents the server configuration structure
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// ReadConfig reads and unmarshal the YAML configuration from the given file
func ReadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// @title Touchly API
// @version 1.0
// @description API Documentation for the Touchly Backend

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
// @tokenUrl http://localhost:8080/auth
// @description This API uses JWT Bearer token. You can get one at /auth
func main() {
	configPath := flag.String("config", "config.yaml", "Path to the config file")
	flag.Parse()

	config, err := ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	storage, err := db.NewDB(config.DbConnString)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}

	defer storage.Close()

	// Create a new Echo instance
	r := chi.NewRouter()

	// Create a new API instance

	email := services.NewEmailClient(config.ResendApiKey)

	api := admSvc.NewApi(storage, email)

	app := transport.New(api, config.JWTSecret)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition"
	))

	app.RegisterRoutes(r)

	// Start the server
	log.Printf("Starting server on %s:%s\n", config.Server.Host, config.Server.Port)

	if err := http.ListenAndServe(config.Server.Host+":"+config.Server.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
