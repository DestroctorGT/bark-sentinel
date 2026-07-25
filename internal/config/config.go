// Package config carga y valida la configuración de la aplicación desde variables de entorno.
package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	WhatsAppToken       string
	WhatsAppPhoneNumber string
	ButcherPhoneNumber  string
	GmailUser           string
	GmailPassword       string
	SMTPHost            string
	SMTPPort            string
}

// Load carga los archivos .env y devuelve la configuracion validada.
// El archivo más específico (.env.local) sobreescribe al base (.env).
func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "local")

	_ = godotenv.Load(".env." + appEnv)
	_ = godotenv.Load(".env")

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV", "local"),
		WhatsAppToken:       os.Getenv("WHATSAPP_TOKEN"),
		WhatsAppPhoneNumber: os.Getenv("WHATSAPP_PHONE_NUMBER"),
		ButcherPhoneNumber:  os.Getenv("BUTCHER_PHONE_NUMBER"),
		GmailUser:           os.Getenv("GMAIL_USER"),
		GmailPassword:       os.Getenv("GMAIL_PASSWORD"),
		SMTPHost:            getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:            getEnv("SMTP_PORT", "587"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.WhatsAppToken == "" {
		return errors.New("WhatsAppToken es requerido")
	}
	if c.WhatsAppPhoneNumber == "" {
		return errors.New("WhatsAppPhoneNumber es requerido")
	}
	if c.ButcherPhoneNumber == "" {
		return errors.New("ButcherPhoneNumber es requerido")
	}
	if c.GmailUser == "" {
		return errors.New("GmailUser es requerido")
	}
	if c.GmailPassword == "" {
		return errors.New("GmailPassword es requerido")
	}
	if c.SMTPHost == "" {
		return errors.New("SmtpHost es requerido")
	}
	if c.SMTPPort == "" {
		return errors.New("SmtpPort es requerido")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultVal
}
