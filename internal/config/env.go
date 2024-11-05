package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/danielmesquitta/openai-web-scraper/internal/pkg/validator"
	"github.com/spf13/viper"
)

type Env struct {
	v validator.Validator

	ChromePath     string `mapstructure:"CHROME_PATH"      validate:"required"`
	CDPPort        string `mapstructure:"CDP_PORT"         validate:"required"`
	DataFolderPath string `mapstructure:"DATA_FOLDER_PATH" validate:"required"`

	CuriositiesFolderPath string
	QuizzesFolderPath     string
	Breeds                []string
}

func LoadEnv(v validator.Validator) *Env {
	e := &Env{
		v: v,
	}

	if err := e.loadEnv(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	if err := e.setDefaults(); err != nil {
		log.Fatalf("failed to set defaults: %v", err)
	}

	if err := e.v.Validate(e); err != nil {
		log.Fatalf("failed to validate config: %v", err)
	}

	return e
}

func (e *Env) loadEnv() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&e); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func (e *Env) setDefaults() error {
	e.CuriositiesFolderPath = path.Join(e.DataFolderPath, "curiosities")
	if _, err := os.Stat(e.CuriositiesFolderPath); err != nil &&
		os.IsNotExist(err) {
		if err := os.Mkdir(e.CuriositiesFolderPath, 0755); err != nil {
			return fmt.Errorf("failed to create curiosities folder: %w", err)
		}
	}

	e.QuizzesFolderPath = path.Join(e.DataFolderPath, "quizzes")
	if _, err := os.Stat(e.QuizzesFolderPath); err != nil &&
		os.IsNotExist(err) {
		if err := os.Mkdir(e.QuizzesFolderPath, 0755); err != nil {
			return fmt.Errorf("failed to create quizzes folder: %w", err)
		}
	}

	e.Breeds = breeds

	return nil
}
