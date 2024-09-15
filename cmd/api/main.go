package main

import (
	"TriceraPass/cmd/api/auth"
	"TriceraPass/internal/repositories"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {

}

// Main Application Config

type application struct {
	DSN          string
	Domain       string
	Repository   *repositories.GORMRepo
	auth         auth.Auth
	JWTSecret    string
	JWTAudience  string
	JWTIssuer    string
	CookieDomain string
	// APIKey       string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Can not locate the env file: %v", err))
	}

	var app application

	// Load the settings
	confFile := os.Getenv("CONFIG_FILE")
	if confFile == "" {
		confFile = "./settings.yml"
	}

	config, err := loadConfig(confFile)
	if err != nil {
		log.Fatal(fmt.Printf("Error loading configuration: %v", err))
	}

	defaultDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s connect_timeout=%s",
		os.Getenv("POSTGRES_HOST"),
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSL,
		config.Database.Timezone,
		config.Database.ConnectTimout,
	)

	// read from command line
	flag.StringVar(&app.DSN, "dsn", defaultDSN, "Postgres connection string")
	flag.StringVar(&app.JWTSecret, "jwt-secret", config.Security.JWT.Secret, "JWT signing secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", config.Security.JWT.Issuer, "JWT signing issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", config.Security.JWT.Audience, "JWT signing audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", config.Application.CookieDomain, "Cookie domain")
	flag.StringVar(&app.Domain, "domain", config.Application.Domain, "Application domain")
	flag.Parse()

	// Log the final config values
	log.Printf("App Domain: %s", app.Domain)
	log.Printf("Development mode: %v", config.Server.DevelopmentMode)

	// Initialize the repository
	app.Repository = &repositories.GORMRepo{}

	// Initiate the auth object
	app.auth = auth.Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "refresh_token",
		CookieDomain:  app.CookieDomain,
	}

	// Open database

	app.Repository.DB, err = gorm.Open(postgres.Open(app.DSN), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Migrate the Schema
	err = app.Repository.Migrate()
	if err != nil {
		log.Println(err)
		return
	}

	fs := http.FileServer(http.Dir("./docs/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	// Handle the home route
	// http.HandleFunc("/", app.Home)

	// Starting the webserver
	log.Printf("Starting the application on port: %v", config.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port), app.routes())
	if err != nil {
		log.Fatal(err)
		return
	}
}
