package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Home returns an HTTP handler function that serves the home page or API status information.
// It dynamically loads configuration and presents either an HTML page or a JSON response based on the Accept header.
//
// Parameters:
// - app: A pointer to the application context containing configuration and services.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function.
func Home(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load environment variables
		err := godotenv.Load()
		if err != nil {
			log.Fatal(fmt.Printf("Cannot locate the env file: %v", err))
		}

		// Define a struct to hold the response data
		type Info struct {
			API struct {
				Name        string        `json:"name"`
				Version     string        `json:"version"`
				Description template.HTML `json:"description"`
			} `json:"api"`
			Application struct {
				ClientName string `json:"client_name"`
				Domain     string `json:"domain"`
			} `json:"application"`
			Styles struct {
				HeaderBackground string
				HeaderColor      string
				HeaderFont       string
				BodyFont         string
				BodyColor        string
				BodyBackground   string
				HeaderFontSize   string
			}
		}

		// Load configuration from the config file
		confFile := os.Getenv("CONFIG_FILE")
		conf, err := application.LoadConfig(confFile)
		if err != nil {
			http.Error(w, "Could not load configuration file", http.StatusInternalServerError)
			return
		}

		description := template.HTML(strings.ReplaceAll(conf.API.Description, "!", "! <br>"))

		info := &Info{
			API: struct {
				Name        string        `json:"name"`
				Version     string        `json:"version"`
				Description template.HTML `json:"description"`
			}{
				Name:        conf.API.Name,
				Version:     conf.API.Version,
				Description: description,
			},
			Application: struct {
				ClientName string `json:"client_name"`
				Domain     string `json:"domain"`
			}{
				ClientName: conf.Application.ClientName,
				Domain:     conf.Application.Domain,
			},
			Styles: struct {
				HeaderBackground string
				HeaderColor      string
				HeaderFont       string
				BodyFont         string
				BodyColor        string
				BodyBackground   string
				HeaderFontSize   string
			}{
				HeaderBackground: conf.Styles.HeaderBackground,
				HeaderColor:      conf.Styles.HeaderColor,
				HeaderFont:       conf.Styles.HeaderFont,
				BodyFont:         conf.Styles.BodyFont,
				BodyColor:        conf.Styles.BodyColor,
				BodyBackground:   conf.Styles.BodyBackground,
				HeaderFontSize:   conf.Styles.HeaderFontSize,
			},
		}

		// Check the Accept header and return either JSON or HTML
		acceptHeader := r.Header.Get("Accept")

		if strings.Contains(acceptHeader, "application/json") {
			payload := struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Info    *Info  `json:"info"`
			}{
				Status:  "active",
				Message: "Authentication service is up and running!",
				Info:    info,
			}

			err = utils.WriteJSON(w, http.StatusOK, payload)
			if err != nil {
				http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
			}
		} else {
			tmpl, err := template.ParseFiles("./template/index.html")
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
				return
			}

			err = tmpl.Execute(w, info)
			if err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}
