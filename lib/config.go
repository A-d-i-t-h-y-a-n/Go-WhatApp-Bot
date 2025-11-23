package lib

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
	"github.com/joho/godotenv"
)

type Configuration struct {
	HANDLERS  string `json:"handlers"`
	SUDO      string `json:"sudo"`
	MODE      string `json:"mode"`
	LOG_MSG   bool   `json:"log_msg"`
	READ_MSG  bool   `json:"read_msg"`
	READ_CMD  bool   `json:"read_cmd"`
	ERROR_MSG bool   `json:"error_msg"`
	QR        bool   `json:"qr"`
	PORT      string `json:"port"`
}

var (
	Config    Configuration
	Client    *whatsmeow.Client
	StartTime time.Time
)

func LoadConfig() {
	_ = godotenv.Load()

	portFlag := flag.String("port", "", "Port number")
	publicFlag := flag.Bool("public", false, "Run in public mode")
	configFileFlag := flag.String("config", "", "Config file path")

	flag.Parse()

	Config = Configuration{
		HANDLERS:  getEnv("HANDLERS", "."),
		SUDO:      getEnv("SUDO", ""),
		MODE:      getEnv("MODE", "public"),
		LOG_MSG:   getEnvBool("LOG_MSG", true),
		READ_MSG:  getEnvBool("READ_MSG", true),
		READ_CMD:  getEnvBool("READ_CMD", true),
		ERROR_MSG: getEnvBool("ERROR_MSG", true),
		QR:        getEnvBool("QR", true),
		PORT:      getEnv("PORT", "8080"),
	}

	if *configFileFlag != "" {
		if data, err := os.ReadFile(*configFileFlag); err == nil {
			_ = json.Unmarshal(data, &Config)
		}
	}

	if *portFlag != "" {
		Config.PORT = *portFlag
	}

	if *publicFlag {
		Config.MODE = "public"
	}
}
