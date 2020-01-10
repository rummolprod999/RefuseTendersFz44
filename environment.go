package main

import (
	"database/sql"
	"fmt"
	"github.com/buger/jsonparser"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type Filelog string

var DirLog = "log_refuse44"
var SetFile = "settings.json"
var FileLog Filelog
var mutex sync.Mutex
var BotToken string
var ChannelId int64
var Dsn string
var CountPage = 19
var FileDB = "bd_purchase.sqlite"
var StartUrl = ""

func DbConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=OFF&_synchronous=OFF", FileDB))
	return db, err
}

func CreateLogFile() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dirlog := filepath.FromSlash(fmt.Sprintf("%s/%s", dir, DirLog))
	if _, err := os.Stat(dirlog); os.IsNotExist(err) {
		err := os.MkdirAll(dirlog, 0711)

		if err != nil {
			fmt.Println("Не могу создать папку для лога")
			os.Exit(1)
		}
	}
	t := time.Now()
	ft := t.Format("2006-01-02")
	FileLog = Filelog(filepath.FromSlash(fmt.Sprintf("%s/log_refuse44_%v.log", dirlog, ft)))
}

func Logging(args ...interface{}) {
	mutex.Lock()
	file, err := os.OpenFile(string(FileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println("Ошибка записи в файл лога", err)
		return
	}
	fmt.Fprintf(file, "%v  ", time.Now())
	for _, v := range args {

		fmt.Fprintf(file, " %v", v)
	}
	//fmt.Fprintf(file, " %s", UrlXml)
	fmt.Fprintln(file, "")
	mutex.Unlock()
}

func ReadSetting() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filetemp := filepath.FromSlash(fmt.Sprintf("%s/%s", dir, SetFile))
	file, err := os.Open(filetemp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	BotToken, err = jsonparser.GetString(b, "bot_token")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ChannelId, err = jsonparser.GetInt(b, "channel_id")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	startUrl, err := jsonparser.GetString(b, "url")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	DbName, err := jsonparser.GetString(b, "db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	UserDb, err := jsonparser.GetString(b, "user_db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	PassDb, err := jsonparser.GetString(b, "pass_db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dsn = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=true&readTimeout=60m&maxAllowedPacket=0&timeout=60m&writeTimeout=60m&autocommit=true&loc=Local", UserDb, PassDb, DbName)
	re := regexp.MustCompile(`&pageNumber=\d{1,2}`)
	startUrl = re.ReplaceAllString(startUrl, "")
	StartUrl = fmt.Sprintf("%s&pageNumber=", startUrl)
	if BotToken == "" || ChannelId == 0 || StartUrl == "" {
		fmt.Println("Check file with settings")
		os.Exit(1)
	}
}

func CreateNewDB() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fileDB := filepath.FromSlash(fmt.Sprintf("%s/%s", dir, FileDB))
	if _, err := os.Stat(fileDB); os.IsNotExist(err) {
		Logging(err)
		f, err := os.Create(fileDB)
		if err != nil {
			Logging(err)
			panic(err)
		}
		err = f.Chmod(0777)
		if err != nil {
			Logging(err)
			//panic(err)
		}
		err = f.Close()
		if err != nil {
			Logging(err)
			panic(err)
		}
		db, err := DbConnection()
		if err != nil {
			Logging(err)
			panic(err)
		}
		defer db.Close()
		_, err = db.Exec(`CREATE TABLE "purchases" (
	"Id"	INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
	"purchase_number"	TEXT,
	"date_update"	TEXT
)`)
		if err != nil {
			Logging(err)
			panic(err)
		}
		_, err = db.Exec(`CREATE INDEX "pur_num" ON "purchases" (
	"purchase_number",
	"date_update"
)`)
		if err != nil {
			Logging(err)
			panic(err)
		}
	}
}

func CreateEnv() {
	ReadSetting()
	CreateLogFile()
	CreateNewDB()
}
