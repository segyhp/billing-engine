package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AppConfig struct {
	Environment              string  `mapstructure:"environment"`
	LogLevel                 string  `mapstructure:"log_level"`
	LoanAmount               float64 `mapstructure:"loan_amount"`
	LoanDurationWeeks        int     `mapstructure:"loan_duration_weeks"`
	AnnualInterestRate       float64 `mapstructure:"annual_interest_rate"`
	DelinquentWeeksThreshold int     `mapstructure:"delinquent_weeks_threshold"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Bind environment variables
	bindEnvVars()

	viper.AutomaticEnv()

	// Try to read config file, but don't fail if it doesn't exist
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "billing_user")
	viper.SetDefault("database.password", "billing_pass")
	viper.SetDefault("database.name", "billing_engine")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "300s")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// App defaults
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "debug")
	viper.SetDefault("app.loan_amount", 5000000.0)
	viper.SetDefault("app.loan_duration_weeks", 50)
	viper.SetDefault("app.annual_interest_rate", 0.10)
	viper.SetDefault("app.delinquent_weeks_threshold", 2)
}

func bindEnvVars() {
	// Server
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// App
	viper.BindEnv("app.environment", "APP_ENV")
	viper.BindEnv("app.log_level", "LOG_LEVEL")
	viper.BindEnv("app.loan_amount", "LOAN_AMOUNT")
	viper.BindEnv("app.loan_duration_weeks", "LOAN_DURATION_WEEKS")
	viper.BindEnv("app.annual_interest_rate", "ANNUAL_INTEREST_RATE")
	viper.BindEnv("app.delinquent_weeks_threshold", "DELINQUENT_WEEKS_THRESHOLD")
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}
