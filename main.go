package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"

	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mpuzanov/bill18test/config"
	"github.com/mpuzanov/bill18test/mail"
	"github.com/mpuzanov/bill18test/models"
)

const (
	statfile    = "web_check.log"
	logFileName = "check.log"
)

var (
	urlsTest      []models.UrlsTest
	urlsResponse  []models.URLResponseHistory
	urls          models.UrlsTest
	cfg           *config.Config
	configModtime int64
	// Лог в файл
	logger *log.Logger
)

//var impl = template.Must(template.ParseFiles("templates/history.html")) // для разбора шаблона 1 раз при запуске сервиса

var configFileName = flag.String("conf", "config.yaml", "config filename")

func main() {
	flag.Parse()

	logFile, _ := os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	defer logFile.Close()
	//logger = log.New(os.Stdout, "info ", log.LstdFlags|log.Lshortfile)
	logger = log.New(logFile, "INFO: ", log.LstdFlags|log.Lshortfile)

	if *configFileName == "" {
		log.Fatal("Конфиг-файл не задан!")
	}
	if _, err := os.Stat(*configFileName); os.IsNotExist(err) {
		log.Fatalf("Конфиг-файл %s не найден!", *configFileName)
	}

	//загружаем конфиг (далее он периодически проверяется в процедуре checkLoop)
	var err error
	cfg, err = reloadConfig(*configFileName)
	checkErr(err)
	//logger.Printf("%#v", cfg)

	go checkLoop() // Проверяем доступность сайтов

	fmt.Printf("Listening on port :%d\n", cfg.Port)
	http.HandleFunc("/", indexHandler)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Выдает историю проверок в браузер
	impl := template.Must(template.ParseFiles("templates/history.html"))
	impl.Execute(w, urlsResponse)
}

func saveHistory(s models.URLResponseHistory) {
	// добавляет запись в массив с историей проверок
	urlsResponse = append(urlsResponse, s)
	if len(urlsResponse) > cfg.HistLength {
		urlsResponse = urlsResponse[1:]
	}
}

// checkLoop функция проверки сайтов по заданному списку
func checkLoop() {
	for {
		var textToSendMail = "Ошибки тестирования сайтов: \n\n"
		isErrorTest := false // Признак наличия ошибочного теста из группы
		for _, url := range urlsTest {
			resTest, objResponse := runCheck(url)
			msg := fmt.Sprintf("%s; %s; %s; %s", objResponse.Name, objResponse.URI, objResponse.Time, objResponse.Status)
			logToFile(msg)
			saveHistory(objResponse)
			if !resTest {
				isErrorTest = true
				textToSendMail += msg + "\n" // добавляем в текст письма при ошибке
			}
		}
		if isErrorTest && cfg.ErrorSendEmail {
			logger.Printf("Отправляем на адрес: %s сообщение:\n%s\n", cfg.ToEmail, textToSendMail)
			if err := mail.SendEmail(cfg, "bill18test", cfg.ToEmail, "Ошибка проверки сайтов", textToSendMail, ""); err != nil {
				logger.Println(err)
			}
		}

		_, err := reloadConfig(*configFileName) // Считываем конфиг (вдруг добавили ещё сайты для проверки)
		checkErr(err)

		time.Sleep(time.Duration(cfg.Timeout) * time.Minute) // Ждём заданное количество времени
	}
}

// logToFile Сохраняем строку в файл
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
func reloadConfig(configName string) (cfg *config.Config, err error) {
	logger.Printf("Проверяем конфигурационный файл: %s ", configName)
	info, err := os.Stat(configName)
	if err != nil {
		return nil, err
	}
	if configModtime != info.ModTime().UnixNano() {
		configModtime = info.ModTime().UnixNano()
		cfg, err = config.ReadConfig(configName)
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
				u.URI, u.Site = gerURLAndParams(url.HTTProtocol, url.Hostapi, p.Path, p.Params)
				u.BasicAuth.Username = url.BasicAuth.Username
				u.BasicAuth.Password = url.BasicAuth.Password
				urlsTest = append(urlsTest, u)
			}
		}
		//log.Println(urlsTest)
		return cfg, nil
	}
	logger.Println("Файл не изменился")

	return
}

// runCheck Проверяем сайт
func runCheck(url models.UrlsTest) (bool, models.URLResponseHistory) {
	// возвращает true — если сервис доступен, false, если нет и текст сообщения
	tm := time.Now().Format("2006–01–02 15:04:05")
	client := &http.Client{Timeout: time.Second * 15} // Создаём своего клиента

	uri := url.URI
	logger.Println("Адрес с учётом параметров: ", uri)

	req, _ := http.NewRequest("GET", uri, nil)
	if url.BasicAuth.Username != "" {
		req.Header.Add("Authorization", "Basic "+basicAuth(url.BasicAuth.Username, url.BasicAuth.Password))
	}

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
func gerURLAndParams(HTTProtocol, hostapi, path string, paramsJSON map[string]string) (u, site string) {
	params := url.Values{}
	for key, value := range paramsJSON {
		//log.Println("key : ", key, " value : ", value)
		params.Add(key, value)
	}
	uri := &url.URL{
		Scheme: HTTProtocol,
		Host:   hostapi,
		Path:   path,
	}
	site = uri.String()
	uri.RawQuery = params.Encode()
	//log.Println(uri.String(), site)
	return uri.String(), site
}
