package models

import (
	"encoding/json"
	"fmt"
)

// HTTPBasicAuthenticator Структура для доступа к сайту
type HTTPBasicAuthenticator struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

//UrlsTest структура для тестирования сайтов
type UrlsTest struct {
	Name string `yaml:"name"`
	// Сайт без параметров
	Site   string            `yaml:"site,omitempty"`
	Path   string            `yaml:"path,omitempty"`
	Params map[string]string `yaml:"params,omitempty"`
	// Сайт с параметрами
	URI       string                 `yaml:"uri,omitempty"`
	BasicAuth HTTPBasicAuthenticator `yaml:"HTTPBasicAuthenticator,omitempty"`
}

//URLResponseHistory Результаты проверки статуса сайта
type URLResponseHistory struct {
	UrlsTest
	Time   string
	Status string
}

//GetParams Выдаём параметры в виде строки Ключ=Значение
func (u *UrlsTest) GetParams() string {
	if len(u.Params) == 0 {
		return ""
	}
	s := ""
	for key, value := range u.Params {
		s += fmt.Sprintf("%s=%s,", key, value)
	}
	return s[:len(s)-1]
}

//GetParamsJSON Выдаём параметры в формате JSON
func (u *UrlsTest) GetParamsJSON() string {
	if len(u.Params) == 0 {
		return ""
	}
	mapVar, _ := json.Marshal(u.Params)
	return string(mapVar)
}
