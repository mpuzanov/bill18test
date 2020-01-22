package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/mpuzanov/bill18test/internal/app/models"
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

//EmailCredentials Структура настройки сервера smtp
type EmailCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
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

//ReadConfig Читаем конфигурацию из файла
func ReadConfig(configName string) (x *Config, err error) {
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
