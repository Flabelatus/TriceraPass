package main

// Config schema from settings.yml
type Config struct {
	API struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
		RateLimit   struct {
			RequestPerMinute int `yaml:"requestsPerMinute"`
			burst            int `yaml:"burst"`
		} `yaml:"rateLimiting"`
	} `yaml:"api"`

	Server struct {
		Port            int    `yaml:"port"`
		Host            string `yaml:"host"`
		DevelopmentMode bool   `yaml:"developmentMode"`
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
			JWTSecret   string `yaml:"secret"`
			JWTIssuer   string `yaml:"issuer"`
			JWTAudience string `yaml:"audience"`
		} `yaml:"jwt"`
	} `yaml:"security"`

	Application struct {
		ClientName   string `yaml:"clientName"`
		CookieDomain string `yaml:"cookieDomain"`
		Domain       string `yaml:"domain"`
	} `yaml:"application"`
	Styles struct {
		HeaderBackground string `yaml:"headerBackground"`
		HeaderColor      string `yaml:"headerColor"`
		HeaderFont       string `yaml:"headerFont"`
		BodyFont         string `yaml:"bodyFont"`
		BodyColor        string `yaml:"bodyColor"`
		BodyBackground   string `yaml:"bodyBackground"`
		HeaderFontSize   string `yaml:"headerFontSize"`
	} `yaml:"styles"`
}
