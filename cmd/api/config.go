package main

// Config schema from settings.yml
type Config struct {
	API struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
		RateLimit   struct {
			RequestsPerMinute int `yaml:"requests_per_minute"`
			Burst             int `yaml:"burst"`
		} `yaml:"rate_limiting"`
	} `yaml:"api"`

	Server struct {
		Port            int    `yaml:"port"`
		Host            string `yaml:"host"`
		DevelopmentMode bool   `yaml:"development_mode"`
	} `yaml:"server"`

	Database struct {
		User          string `yaml:"user"`
		Password      string `yaml:"password"`
		Name          string `yaml:"dbname"`
		Host          string `yaml:"host"`
		Port          string `yaml:"port"`
		SSL           string `yaml:"sslmode"`
		Timezone      string `yaml:"timezone"`
		ConnectTimout string `yaml:"connect_timeout"`
	} `yaml:"database"`

	Security struct {
		JWT struct {
			Secret   string `yaml:"secret"`
			Issuer   string `yaml:"issuer"`
			Audience string `yaml:"audience"`
		} `yaml:"jwt"`
	} `yaml:"security"`

	Application struct {
		ClientName   string `yaml:"client_name"`
		CookieDomain string `yaml:"cookie_domain"`
		Domain       string `yaml:"domain"`
	} `yaml:"application"`

	Styles struct {
		HeaderBackground string `yaml:"header_background"`
		HeaderColor      string `yaml:"header_color"`
		HeaderFont       string `yaml:"header_font"`
		BodyFont         string `yaml:"body_font"`
		BodyColor        string `yaml:"body_color"`
		BodyBackground   string `yaml:"body_background"`
		HeaderFontSize   string `yaml:"header_font_size"`
	} `yaml:"styles"`
}
