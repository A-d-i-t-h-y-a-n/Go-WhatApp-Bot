package lib

import (
	"time"
	"flag"
	"os"
	"encoding/json"

	"go.mau.fi/whatsmeow"
	"github.com/joho/godotenv"
)

type Configuration struct {
	HANDLERS  string
	SUDO      string
	MODE      string
	LOG_MSG   bool
	READ_MSG  bool
	READ_CMD  bool
	ERROR_MSG bool
	QR        bool
}

var (
	Config    Configuration
	Client    *whatsmeow.Client
	StartTime time.Time
)

func LoadConfig() {
	_ = godotenv.Load()

	configFile := flag.String("config", "", "")
	flag.Parse()

	Config = Configuration{
		HANDLERS:  getEnv("HANDLERS", "."),
		SUDO:      getEnv("SUDO", ""),
		MODE:      getEnv("MODE", "public"),
		LOG_MSG:   getEnvBool("LOG_MSG", true),
		READ_MSG:  getEnvBool("READ_MSG", true),
		READ_CMD:  getEnvBool("READ_CMD", true),
		ERROR_MSG: getEnvBool("ERROR_MSG", true),
		QR:        getEnvBool("QR", false),
	}

	if *configFile != "" {
		data, err := os.ReadFile(*configFile)
		if err == nil {
			_ = json.Unmarshal(data, &Config)
		}
	}
}