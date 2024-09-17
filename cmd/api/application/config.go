package application

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config defines the schema for the settings.yml file.
// It holds configuration settings for the API, server, database, security, application, and styles.
type Config struct {
	API struct {
		Name        string `yaml:"name"`        // API name
		Version     string `yaml:"version"`     // API version
		Description string `yaml:"description"` // API description
		RateLimit   struct {
			RequestsPerMinute int `yaml:"requests_per_minute"` // Max requests per minute
			Burst             int `yaml:"burst"`               // Burst size for rate limiting
		} `yaml:"rate_limiting"` // Rate limiting configuration
	} `yaml:"api"`

	Server struct {
		Port            int    `yaml:"port"`             // Server port number
		Host            string `yaml:"host"`             // Server host address
		DevelopmentMode bool   `yaml:"development_mode"` // Is the server in development mode
	} `yaml:"server"`

	Database struct {
		User          string `yaml:"user"`            // Database username
		Password      string `yaml:"password"`        // Database password
		Name          string `yaml:"dbname"`          // Database name
		Host          string `yaml:"host"`            // Database host address
		Port          string `yaml:"port"`            // Database port
		SSL           string `yaml:"sslmode"`         // SSL mode for database connection
		Timezone      string `yaml:"timezone"`        // Database timezone
		ConnectTimout string `yaml:"connect_timeout"` // Database connection timeout
	} `yaml:"database"`

	Security struct {
		JWT struct {
			Secret   string `yaml:"secret"`   // JWT secret for signing tokens
			Issuer   string `yaml:"issuer"`   // JWT issuer claim
			Audience string `yaml:"audience"` // JWT audience claim
		} `yaml:"jwt"`
	} `yaml:"security"`

	Application struct {
		ClientName   string `yaml:"client_name"`   // Name of the client application
		CookieDomain string `yaml:"cookie_domain"` // Domain for setting cookies
		Domain       string `yaml:"domain"`        // Application domain
	} `yaml:"application"`

	Styles struct {
		HeaderBackground string `yaml:"header_background"` // Background color for the header
		HeaderColor      string `yaml:"header_color"`      // Text color for the header
		HeaderFont       string `yaml:"header_font"`       // Font used for the header
		BodyFont         string `yaml:"body_font"`         // Font used for the body
		BodyColor        string `yaml:"body_color"`        // Text color for the body
		BodyBackground   string `yaml:"body_background"`   // Background color for the body
		HeaderFontSize   string `yaml:"header_font_size"`  // Font size for the header
	} `yaml:"styles"`
}

// LoadConfig loads and parses the settings.yml file into the Config struct.
// It reads the configuration file and unmarshals its YAML content into the Config struct.
//
// Parameters:
// - configFile: The path to the configuration file (settings.yml).
//
// Returns:
// - *Config: A pointer to the parsed Config struct.
// - error: An error if the file cannot be read or parsed.
func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return config, nil
}

// isOriginAllowed checks if a given origin is in the list of allowed origins for CORS.
//
// Parameters:
// - origin: The origin to check.
// - allowedOrigins: A slice of allowed origins.
//
// Returns:
// - bool: True if the origin is allowed, false otherwise.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
