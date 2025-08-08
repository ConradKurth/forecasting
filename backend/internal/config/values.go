package config

type serviceConfig struct {
	Env        string `env:"GO_ENV"`
	Service    service
	Database   database
	Redis      redis
	Shopify    shopify
	Frontend   frontend
	CORS       cors
	Encryption encryption
	Logging    logging
}

type service struct {
	Env string `long:"env" env:"SERVICE_ENV" description:"Service environment"`
}

type frontend struct {
	URL string `long:"url" default:"" env:"FRONTEND_URL" description:"Frontend URL"`
}

type shopify struct {
	ClientID     string   `long:"client-id" default:"" env:"SHOPIFY_CLIENT_ID" description:"Shopify Client ID"`
	ClientSecret string   `long:"client-secret" default:"" env:"SHOPIFY_CLIENT_SECRET" description:"Shopify Client Secret"`
	RedirectURL  string   `long:"redirect-url" default:"" env:"SHOPIFY_REDIRECT_URL" description:"Shopify Redirect URL"`
	Scopes       []string `long:"scopes" default:"read_products,read_locations,read_inventory,read_orders" env:"SHOPIFY_SCOPES" description:"Shopify Scopes"`
}

type cors struct {
	AllowedOrigins []string `long:"allowed-origins" env-delim:"," default:"http://localhost:5173" env:"CORS_ALLOWED_ORIGINS" description:"CORS Allowed Origins"`
}

type database struct {
	URL string `long:"database-url" env:"DATABASE_URL" description:"Database connection URL" required:"true"`
}

type redis struct {
	URL string `long:"redis-url" env:"REDIS_URL" default:"localhost:6379" description:"Redis connection URL"`
}

type encryption struct {
	SecretKey string `long:"secret-key" env:"SECRET_KEY" description:"32-byte secret key for AES-256-GCM encryption" required:"true"`
}

type logging struct {
	Level string `long:"log-level" env:"LOG_LEVEL" default:"info" description:"Log level (debug, info, warn, error)"`
}
