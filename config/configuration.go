package config

import (
	"github.com/spf13/viper"
	"github.com/maxime1907/crudify/logger"
)

type DBInfo struct {
	Host     string
	User     string
	Dbname   string
	Password string
	Sslmode  string
	Driver   string
}

type RouterInfo struct {
	Username string
	Password string
	Port int
}

type UpdaterInfo struct {
	Url      string
	Interval int
}

type CORSInfo struct {
	Origins []string
	Methods []string
	Headers []string
}

type TLSInfo struct {
	Crt string
	Key string
}

type SMTPInfo struct {
	OwnerEmail string
	Password string
	Host string
	Port int
}

type Config struct {
	Database	DBInfo
	Server		RouterInfo
	Update		UpdaterInfo
	Cors		CORSInfo
	TLS			TLSInfo
	SMTP		SMTPInfo
}

var config Config

func Get() Config {
	return config
}

// Read a given config file and returns its result in a Config struct
func Read(filename string, path string) (Config, error) {

	viper.SetConfigName(filename)
	viper.AddConfigPath(path)
	return _readAndSetConf()
}

// Read a given config file and returns its result in a Config struct
func ReadFullPath(fullPath string) (Config, error) {

	logger.Log(nil).Debug().Msg("Reading configuration file " + fullPath)

	viper.SetConfigFile(fullPath)
	return _readAndSetConf()
}

func _readAndSetConf() (configuration Config, err error) {

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&configuration)
	config = configuration
	return configuration, err
}
