package settings

import (
	"encoding/json"
	"os"
)

// GetSettings attempts to retrieve
// application settings from a config file
func GetSettings() (p PaoSettings, e error) {
	p = PaoSettings{}
	path := os.Getenv("PAO_CONF")
	if path == "" {
		path = "conf/paoSettings.json"
	}
	file, err := os.Open(path)
	if err != nil {
		e = err
		return
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&p)
	if err != nil {
		e = err
	}
	return
}

type dbConfig struct {
	Driver, ConnectionString string
}
type authConfig struct {
	EncryptionKey string
}

// AiConfig contains information necessary to connect to an external AI server
type AiConfig struct {
	Name    string
	Address string
}

// PaoSettings for pao as specified
// in conf/paoSettings.json
type PaoSettings struct {
	DbConfig   dbConfig
	AuthConfig authConfig
	Ais        []AiConfig
}
