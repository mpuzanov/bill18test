package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"path"

	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mpuzanov/bill18test/models"
)

const (
	configFileName = "configs/config.yaml"
	statfile       = "web_check.log"
	logFileName    = "check.log"
)

var (
	urlsTest     []models.UrlsTest
	urlsResponse []models.URLResponseHistory
	urls         models.UrlsTest
	cfg          *Config
	logger       *log.Logger
)

//var tmpl := template.Must(template.ParseFiles("templates/history.html"))

func main() {
	logFile, _ := os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	defer logFile.Close()
	//logger = log.New(os.Stdout, "info ", log.LstdFlags|log.Lshortfile)
	logger = log.New(logFile, "info ", log.LstdFlags|log.Lshortfile)

	//загружаем конфиг (далее он периодически проверяется в процедуре checkLoop)
	var err error
	cfg, err = reloadConfig(configFileName)
	checkErr(err)
	//logger.Printf("%#v", cfg)

	go checkLoop() // Проверяем доступность сайтов

	fmt.Printf("Listening on port :%d\n", cfg.Port)
	http.HandleFunc("/", indexHandler)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Выдает историю проверок в браузер
	tmpl := template.Must(template.ParseFiles("templates/history.html"))
	tmpl.Execute(w, urlsResponse)
}

func saveHstory(s models.URLResponseHistory) {
	// добавляет запись в массив с историей проверок
	urlsResponse = append(urlsResponse, s)
	if len(urlsResponse) > cfg.HistLength {
		urlsResponse = urlsResponse[1:]
	}
}

// checkLoop фунция проверки сайтов по заданному списку
func checkLoop() {
	var textToSendMail string
	for {
		isSendMail := false
		for _, url := range urlsTest {
			res, objResponse := runCheck(url)
			msg := fmt.Sprintf("%s; %s; %v; %s; %s", objResponse.Name, objResponse.Site, objResponse.GetParamsJSON(), objResponse.Time, objResponse.Status)
			logToFile(msg)
			saveHstory(objResponse)
			if res != true {
				if isSendMail == false {
					isSendMail = true
				}
				textToSendMail += msg // добавляем в текст письма при ошибке
			}
		}
		if isSendMail == true && cfg.ErrorSendEmail == true {
			log.Printf("Отправляем на адрес: %s сообщение: %s\n", cfg.ToEmail, textToSendMail)
			SendEmail(cfg.ToEmail, "Ошибка проверки сайтов", textToSendMail, "")
		}

		_, err := reloadConfig(configFileName) // Считываем конфиг (вдруг добавили ещё сайты для проверки)
		checkErr(err)

		time.Sleep(time.Duration(cfg.Timeout) * time.Minute) // Ждём заданное количество времени
	}
}

// logToFile Сохраняем строку логирования в файл
func logToFile(s string) {
	logger.Printf("Сохраняем строку <%s> в файл %s\n", s, statfile)
	f, err := os.OpenFile(statfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	// Записываем строку в файл
	if _, err = f.WriteString(fmt.Sprintf("%s\n", s)); err != nil {
		logger.Println(err)
	}
}

//checkErr функция обработки ошибок
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//Проверяет время изменения конфигурационного файла
//и перезагружает его если он изменился
//Возвращает errNotModified если изменений нет
func reloadConfig(configName string) (cfg *Config, err error) {
	logger.Printf("Проверяем конфигурационный файл: %s ", configName)
	info, err := os.Stat(configName)
	if err != nil {
		return nil, err
	}
	if configModtime != info.ModTime().UnixNano() {
		configModtime = info.ModTime().UnixNano()
		cfg, err = readConfig(configName)
		if err != nil {
			return nil, err
		}
		logger.Println("reload config parameters")
		urlsResponse = make([]models.URLResponseHistory, 0)
		urlsTest = make([]models.UrlsTest, 0)
		for _, url := range cfg.UrlsTest {
			for _, p := range url.URLParams {
				u := models.UrlsTest{}
				u.Name = p.Name
				u.Params = p.Params
				u.Path = p.Path
				u.Site = path.Join(url.HTTProtocol, url.Hostapi, p.Path)
				u.URI = gerURLAndParams(url.HTTProtocol, url.Hostapi, p.Path, p.Params)
				urlsTest = append(urlsTest, u)
			}
		}
		//fmt.Println(urlsTest)
		return cfg, nil
	}
	logger.Println("Файл неизменился")

	return nil, nil
}

// runCheck Проверяем сайт
func runCheck(url models.UrlsTest) (bool, models.URLResponseHistory) {
	// возвращает true — если сервис доступен, false, если нет и текст сообщения
	tm := time.Now().Format("2006–01–02 15:04:05")
	client := &http.Client{Timeout: time.Second * 10} // Создаём своего клиента

	uri := url.URI
	logger.Println("Адрес с учётом параметров: ", uri)

	req, _ := http.NewRequest("GET", uri, nil)

	req.Header.Add("Authorization", "Basic "+basicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password))

	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error http.Get: %q\n", err)
		return false, models.URLResponseHistory{UrlsTest: url, Time: tm, Status: "Ошибка соединения"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Printf("Сайт %s. Ошибка. http-статус: %s\n", url, resp.Status)
		//log.Printf("*******************************************************************\n")
		//body, _ := ioutil.ReadAll(resp.Body)
		//log.Printf("Сайт %s. Ошибка. http-статус: %s\n", url.Site, string(body))
		//log.Printf("===================================================================\n")
		return false, models.URLResponseHistory{UrlsTest: url, Time: tm, Status: resp.Status}
	}
	//fmt.Sprintf("Сайт %s. Онлайн. http-статус: %d", url, resp.StatusCode)
	return true, models.URLResponseHistory{UrlsTest: url, Time: tm, Status: resp.Status}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// gerURLAndParams Формируем адрес сайта с параметрами
func gerURLAndParams(HTTProtocol, hostapi, path string, paramsJSON map[string]string) string {
	params := url.Values{}
	for key, value := range paramsJSON {
		//log.Println("key : ", key, " value : ", value)
		params.Add(key, value)
	}
	uri := &url.URL{
		Scheme:   HTTProtocol,
		Host:     hostapi,
		Path:     path,
		RawQuery: params.Encode(),
	}
	//fmt.Println(uri.String())
	return uri.String()
}

//func gerURLAndParams(reqURL string, paramsJSON string) string {
// 	if paramsJSON == "" {
// 		return reqURL
// 	}
// 	log.Println(reqURL)
// 	log.Println(paramsJSON)
// 	paramsJSON = strings.Replace(paramsJSON, "'", "\"", -1)
// 	log.Println(paramsJSON)
// 	keyValMap := make(map[string]interface{})
// 	err := json.Unmarshal([]byte(paramsJSON), &keyValMap)
// 	if err != nil {
// 		log.Printf("Error: %q\n", err)
// 	}
// 	params := url.Values{}
// 	for key, value := range keyValMap {
// 		log.Println("key : ", key, " value : ", value)
// 		params.Add(key, value.(string))
// 	}
// 	uri, _ := url.Parse(reqURL)
// 	uri.RawQuery = params.Encode()
// 	reqURL = uri.String()
// 	return reqURL
// }
