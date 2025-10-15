package config

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
}

type JWTConfig struct {
	Token         string        `json:"token"`
	ExpiresIn     time.Duration `json:"expires_in"`
	SigningMethod *jwt.SigningMethodHMAC
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{} // создаем объект конфига
	err = viper.Unmarshal(cfg) // читаем информацию из файла, 
	// конвертируем и затем кладем в нашу переменную cfg
	if err != nil {
		return nil, err
	}

	// Настраиваем JWT
	cfg.setupJWT()

	log.Info("config parsed")

	return cfg, nil
}

func (c *Config) setupJWT() {
	// Получаем JWT секрет из переменных окружения
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-for-flight-api" // ДОЛЖЕН СОВПАДАТЬ
	}

	fmt.Printf("Config - JWT Secret: %s\n", jwtSecret)
	
	c.JWT.Token = jwtSecret
	c.JWT.ExpiresIn = 24 * time.Hour
	c.JWT.SigningMethod = jwt.SigningMethodHS256
}

// Вспомогательная функция для получения переменных окружения
func (c *Config) getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}