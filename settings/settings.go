package settings

import (
	"encoding/json"
	"os"
)

// GetSettings attempts to retrieve
// application settings from a config file
func GetSettings() (p PaoSettings, e error) {
	p = PaoSettings{}
	file, err := os.Open("conf/paoSettings.json")
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

// PaoSettings for pao as specified
// in conf/paoSettings.json
type PaoSettings struct {
	DbConfig   dbConfig
	AuthConfig authConfig
}
