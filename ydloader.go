package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/AlekseiGrigorev/ydloader/internal/config"
	"github.com/AlekseiGrigorev/ydloader/internal/db"
	"github.com/AlekseiGrigorev/ydloader/internal/logger"
	"github.com/AlekseiGrigorev/ydloader/internal/template"
	"github.com/AlekseiGrigorev/ydloader/internal/trace"
	"github.com/AlekseiGrigorev/ydloader/models/integrations"
	"github.com/AlekseiGrigorev/ydloader/models/ydirectlogins"
)

const IntegrationId = 0     //10472 - 50, 7101 - 34
const InputDir = "./input/" //Input data dir (getting from api)
const LogFile = "app.log"

var AppConfig config.Config
var AppDb db.Db
var Log = logger.Log{
	PrintToStdout:   true,
	PrefixDelimiter: " ",
}

type BaseStruct struct {
	Token     string
	Login     string
	Headers   string
	Body      string
	Processed bool
	Started   bool
	NextTry   time.Time
	Try       int
}

type RespStruct struct {
	Status     string
	StatusCode int
	Header     http.Header
	Body       string
}

func main() {
	Log.Log().SetFlags(log.LstdFlags)
	AppConfig = getConfig()

	file, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		Log.Error("Failed to open log file:", err)
		return
	}
	defer file.Close()
	Log.Log().SetOutput(file)
	Log.Info("App started")

	AppDb.Init(AppConfig.Db.Username, AppConfig.Db.Password, AppConfig.Db.Host, AppConfig.Db.Port, AppConfig.Db.Database)

	var structs []*BaseStruct

	if IntegrationId > 0 {
		token, err := getToken(IntegrationId)
		if err != nil {
			Log.Error(err, trace.GetTrace())
			return
		}

		Log.Info(token)

		logins, err := getLogins(IntegrationId)
		if err != nil {
			Log.Error(err, trace.GetTrace())
			return
		}

		for _, login := range logins {
			Log.Info(*login)
		}

		structs, err = fillBaseStructs(token, logins)
		if err != nil {
			Log.Error(err, trace.GetTrace())
			return
		}
	} else {
		logins, err := getAllLogins()
		if err != nil {
			Log.Error(err, trace.GetTrace())
			return
		}

		for _, login := range logins {
			Log.Info(*login)
		}

		structs, err = fillBaseStructsForAllLogins(logins)
		if err != nil {
			Log.Error(err, trace.GetTrace())
			return
		}
	}

	maxFor := 1000
	currFor := 0
	processed := false

	for !processed {
		currFor++
		Log.Info("Cycle " + strconv.Itoa(currFor))
		currProcessed := true
		for i := 0; i < len(structs); i++ {
			if !structs[i].Processed {
				currProcessed = false
			}
			if structs[i].Started || structs[i].Processed {
				continue
			}
			if time.Now().After(structs[i].NextTry) {
				structs[i].Try++
				if structs[i].Try > AppConfig.Http.TryCount {
					structs[i].Processed = true
					continue
				}
				structs[i].Started = true
				go getReportGo(structs[i])
				/*
					err = getReport(structs[i])
					if err != nil {
						structs[i].Processed = true
						fmt.Println(err, trace.GetTrace())
					}
				*/
			}
		}
		processed = currProcessed
		if currFor > maxFor {
			processed = true
		}
		time.Sleep(time.Second * 1)
	}
}

// Process get report as goroutine
func getReportGo(baseStruct *BaseStruct) {
	err := getReport(baseStruct)
	if err != nil {
		baseStruct.Processed = true
		Log.Error(err, trace.GetTrace())
	}
	baseStruct.Started = false
}

// Returns application config struct
func getConfig() config.Config {
	var appConfig config.Config
	viper.SetConfigName("config.yml") // name of config file (without extension)
	viper.SetConfigType("yaml")       // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./config/")  // path to look for the config file in
	err := viper.ReadInConfig()       // Find and read the config file
	if err != nil {                   // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	err = viper.Unmarshal(&appConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}
	return appConfig
}

// Get integration token from DB
func getToken(intId int) (string, error) {
	params := []any{intId}
	tokenModel := integrations.Token{}
	tokenAny, err := AppDb.QueryRow(tokenModel.GetDefaultSql(), params, &tokenModel)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return "", err
	}
	token := tokenAny.(*integrations.Token)
	if token.Token == "" {
		err = errors.New("token not found")
		Log.Error(err, trace.GetTrace())
		return "", err
	}
	fmt.Println(token.Token)
	return token.Token, nil
}

// Get integration logins from DB
func getLogins(intId int) ([]*ydirectlogins.IntegrationLogin, error) {
	params := []any{intId}
	loginModel := ydirectlogins.IntegrationLogin{}
	logins, err := AppDb.Query(loginModel.GetDefaultSql(), params, &loginModel)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	if len(logins) == 0 {
		err = errors.New("logins not found")
		Log.Error(err, trace.GetTrace())
		return nil, err
	}

	return loginModel.ToType(logins), nil
}

// Get all logins from DB
func getAllLogins() ([]*ydirectlogins.AllIntegrationsLogin, error) {
	params := []any{}
	loginModel := ydirectlogins.AllIntegrationsLogin{}
	logins, err := AppDb.Query(loginModel.GetDefaultSql(), params, &loginModel)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	if len(logins) == 0 {
		Log.Error("Logins not found!", trace.GetTrace())
		return nil, err
	}
	return loginModel.ToType(logins), nil
}

// Fill base struct data slice from all logins
func fillBaseStructsForAllLogins(logins []*ydirectlogins.AllIntegrationsLogin) ([]*BaseStruct, error) {
	header := template.TemplateManager{}
	err := header.SetTemplate("./templates/header.json")
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	body := template.TemplateManager{}
	err = body.SetTemplate("./templates/body.json")
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	yesterday := time.Now().Add(-24 * time.Hour)
	headerMap := make(map[string]string)
	headerMap["@AuthorizationToken"] = ""
	headerMap["@Client-Login"] = ""
	bodyMap := make(map[string]string)
	bodyMap["@DateFrom"] = yesterday.Format("2006-01-02")
	bodyMap["@DateTo"] = yesterday.Format("2006-01-02")
	bodyMap["@ReportName"] = ""
	structs := []*BaseStruct{}
	for _, login := range logins {
		headerMap["@AuthorizationToken"] = login.Token
		headerMap["@Client-Login"] = login.Login
		bodyMap["@ReportName"] = strconv.FormatInt(rand.Int63(), 10)
		structs = append(structs, &BaseStruct{
			Token:     login.Token,
			Login:     login.Login,
			Headers:   header.Process(headerMap),
			Body:      body.Process(bodyMap),
			Processed: false,
			Started:   false,
			NextTry:   time.Now().Add(-1 * time.Second),
			Try:       0,
		})
	}
	return structs, nil
}

// Fill base struct data slice for integration logins
func fillBaseStructs(token string, logins []*ydirectlogins.IntegrationLogin) ([]*BaseStruct, error) {
	header := template.TemplateManager{}
	err := header.SetTemplate("./templates/header.json")
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	body := template.TemplateManager{}
	err = body.SetTemplate("./templates/body.json")
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	yesterday := time.Now().Add(-24 * time.Hour)
	headerMap := make(map[string]string)
	headerMap["@AuthorizationToken"] = token
	headerMap["@Client-Login"] = ""
	bodyMap := make(map[string]string)
	bodyMap["@DateFrom"] = yesterday.Format("2006-01-02")
	bodyMap["@DateTo"] = yesterday.Format("2006-01-02")
	bodyMap["@ReportName"] = ""
	structs := []*BaseStruct{}
	for _, login := range logins {
		headerMap["@Client-Login"] = login.Login
		bodyMap["@ReportName"] = strconv.FormatInt(rand.Int63(), 10)
		structs = append(structs, &BaseStruct{
			Login:     login.Login,
			Headers:   header.Process(headerMap),
			Body:      body.Process(bodyMap),
			Processed: false,
			NextTry:   time.Now(),
		})
	}
	return structs, nil
}

// Get report data from YD API
func getReport(baseStruct *BaseStruct) error {
	Log.Info("Get report start", baseStruct.Login)
	path, err := createDir(baseStruct)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	resp, err := post(baseStruct)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	err = writeFileResp(path, resp)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	err = processResp(baseStruct, resp)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	Log.Info("Get report end", baseStruct.Login)
	return nil
}

// Process response data
func processResp(baseStruct *BaseStruct, resp *RespStruct) error {
	switch resp.StatusCode {
	case 200:
		baseStruct.Processed = true
		return nil
	case 201:
		retryin, err := strconv.Atoi(resp.Header.Get("Retryin"))
		if err != nil {
			baseStruct.Processed = true
			Log.Error(err, trace.GetTrace())
			return err
		}
		baseStruct.NextTry = time.Now().Add(time.Duration(retryin) * time.Second)
		return nil
	case 202:
		retryin, err := strconv.Atoi(resp.Header.Get("Retryin"))
		if err != nil {
			baseStruct.Processed = true
			Log.Error(err, trace.GetTrace())
			return err
		}
		baseStruct.NextTry = time.Now().Add(time.Duration(retryin) * time.Second)
		return nil
	case 503:
		baseStruct.NextTry = time.Now().Add(time.Duration(1) * time.Second)
		return nil
	}
	baseStruct.Processed = true
	return nil
}

// Write file with response data
func writeFileResp(path string, resp *RespStruct) error {
	respJson, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	err = writeFile(path, respJson)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	return nil
}

// Write file
func writeFile(path string, content []byte) error {
	filename := filepath.Join(path, time.Now().Format("20060102150405")+".txt")
	err := os.WriteFile(filename, content, 0777)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return err
	}
	return nil
}

// Get data from report service
func post(baseStruct *BaseStruct) (*RespStruct, error) {
	body := bytes.NewBuffer([]byte(baseStruct.Body))
	c := http.Client{Timeout: time.Duration(AppConfig.Http.Timeout) * time.Second}
	req, err := http.NewRequest("POST", AppConfig.Http.ReportsUrl, body)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	headers := map[string]string{}
	err = json.Unmarshal([]byte(baseStruct.Headers), &headers)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.Error(err, trace.GetTrace())
		return nil, err
	}
	return &RespStruct{
		Header:     resp.Header.Clone(),
		Body:       string(respBody),
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
	}, nil
}

// Create input data directory if needed
func createDir(baseStruct *BaseStruct) (string, error) {
	path := InputDir + baseStruct.Login
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(path, 0777)
			if err != nil {
				Log.Error(err, trace.GetTrace())
				return "", err
			}
			return path, nil
		} else {
			Log.Error(err, trace.GetTrace())
			return "", err
		}
	}
	return path, nil
}
