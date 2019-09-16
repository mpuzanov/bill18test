package main

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/mpuzanov/bill18test/models"
)

var (
	configModtime int64
)

//URLParam ...
type URLParam struct {
	Name   string            `yaml:"name,omitempty"`
	Path   string            `yaml:"path,omitempty"`
	Params map[string]string `yaml:"params,omitempty"`
}

// UrlsTestConfig ...
type UrlsTestConfig struct {
	Hostapi     string                        `yaml:"hostapi"`
	HTTProtocol string                        `yaml:"http_protocol,omitempty"`
	URLParams   []URLParam                    `yaml:"url_params,omitempty"`
	BasicAuth   models.HTTPBasicAuthenticator `yaml:"HTTPBasicAuthenticator,omitempty"`
}

// Config - структура для считывания конфигурационного файла
type Config struct {
	LogLevel       string           `yaml:"log_level"`
	Timeout        int              `yaml:"timeout"`
	Port           int              `yaml:"port"`
	HistLength     int              `yaml:"histLength"`
	ToEmail        string           `yaml:"toEmail"`
	UrlsTest       []UrlsTestConfig `yaml:"urlTest"`
	ErrorSendEmail bool             `yaml:"errorSendEmail"`
	SettingsSMTP   EmailCredentials `yaml:"settingsSMTP"`
}

// readConfig Читаем конфигурацию из файла
func readConfig(configName string) (x *Config, err error) {
	logger.Printf("Читаем конфигурацию из файла: %s\n", configName)
	var file []byte
	if file, err = ioutil.ReadFile(configName); err != nil {
		return nil, err
	}
	x = new(Config)
	if err = yaml.Unmarshal(file, x); err != nil {
		return nil, err
	}
	return x, nil
}
