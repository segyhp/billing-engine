package config

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Business  BusinessConfig  `mapstructure:"business"`
	Health    HealthConfig    `mapstructure:"health"`
}

type ServerConfig struct {
	Port string `mapstructure:"SERVER_PORT"`
	Host string `mapstructure:"SERVER_HOST"`
	Env  string `mapstructure:"ENV"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"DATABASE_URL"`
	Host     string `mapstructure:"DATABASE_HOST"`
	Port     string `mapstructure:"DATABASE_PORT"`
	Name     string `mapstructure:"DATABASE_NAME"`
	User     string `mapstructure:"DATABASE_USER"`
	Password string `mapstructure:"DATABASE_PASSWORD"`
}

type RedisConfig struct {
	URL      string `mapstructure:"REDIS_URL"`
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

type SchedulerConfig struct {
	Interval string `mapstructure:"SCHEDULER_INTERVAL"`
	Timezone string `mapstructure:"SCHEDULER_TIMEZONE"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

type BusinessConfig struct {
	DefaultInterestRate  string `mapstructure:"DEFAULT_INTEREST_RATE"`
	DefaultLoanWeeks     int    `mapstructure:"DEFAULT_LOAN_WEEKS"`
	DelinquencyThreshold int    `mapstructure:"DELINQUENCY_THRESHOLD"`
}

type HealthConfig struct {
	Timeout string `mapstructure:"HEALTH_CHECK_TIMEOUT"`
}

// Load reads configuration from environment variables and files
func Load() (*Config, error) {
	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("DEFAULT_INTEREST_RATE", "0.10")
	viper.SetDefault("DEFAULT_LOAN_WEEKS", 50)
	viper.SetDefault("DELINQUENCY_THRESHOLD", 2)
	viper.SetDefault("SCHEDULER_INTERVAL", "24h")
	viper.SetDefault("SCHEDULER_TIMEZONE", "Asia/Jakarta")
	viper.SetDefault("HEALTH_CHECK_TIMEOUT", "5s")

	// Read from environment variables
	viper.AutomaticEnv()

	// Try to read from .env file (optional)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./deployments")

	// Don't fail if .env file doesn't exist
	_ = viper.ReadInConfig()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.Business.DefaultLoanWeeks <= 0 {
		return fmt.Errorf("DEFAULT_LOAN_WEEKS must be greater than 0")
	}

	if c.Business.DelinquencyThreshold <= 0 {
		return fmt.Errorf("DELINQUENCY_THRESHOLD must be greater than 0")
	}

	// Validate interest rate
	if _, err := decimal.NewFromString(c.Business.DefaultInterestRate); err != nil {
		return fmt.Errorf("DEFAULT_INTEREST_RATE must be a valid decimal: %w", err)
	}

	// Validate scheduler interval
	if _, err := time.ParseDuration(c.Scheduler.Interval); err != nil {
		return fmt.Errorf("SCHEDULER_INTERVAL must be a valid duration: %w", err)
	}

	// Validate health check timeout
	if _, err := time.ParseDuration(c.Health.Timeout); err != nil {
		return fmt.Errorf("HEALTH_CHECK_TIMEOUT must be a valid duration: %w", err)
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development" || c.Server.Env == "dev"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production" || c.Server.Env == "prod"
}

// GetDefaultInterestRate returns the default interest rate as decimal
func (c *Config) GetDefaultInterestRate() decimal.Decimal {
	rate, _ := decimal.NewFromString(c.Business.DefaultInterestRate)
	return rate
}

// GetSchedulerInterval returns the scheduler interval as duration
func (c *Config) GetSchedulerInterval() time.Duration {
	duration, _ := time.ParseDuration(c.Scheduler.Interval)
	return duration
}

// GetHealthTimeout returns the health check timeout as duration
func (c *Config) GetHealthTimeout() time.Duration {
	timeout, _ := time.ParseDuration(c.Health.Timeout)
	return timeout
}
