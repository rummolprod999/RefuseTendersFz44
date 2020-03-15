package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

type ParserEisFtp223 struct {
	addDoc  int
	sendDoc int
}

func (t *ParserEisFtp223) run() {
	defer SaveStack()
	currTime := time.Now()
	lastTime := currTime.AddDate(0, 0, -20)
	db, err := sql.Open("mysql", Dsn2)
	if err != nil {
		Logging("error connection to db", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, err := db.Prepare("SELECT prot.protocol_date, prot.purchase_number, prot.url, prot.protocol_date, prot.type_protocol FROM protocols223 prot WHERE prot.protocol_date >= ? AND prot.type_protocol = 'Протокол признания участника уклонившимся от заключения договора'")
	rows, err := stmt.Query(lastTime)
	stmt.Close()
	if err != nil {
		Logging("error execution query", err)
		return
	}
	var protocols = []Puchase{}
	for rows.Next() {
		var purName string
		var purNum string
		var url string
		var pubDate time.Time
		var purchaseDate time.Time
		err = rows.Scan(&pubDate, &purNum, &url, &purchaseDate, &purName)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return
		}
		url = fmt.Sprintf("http://zakupki.gov.ru/223/purchase/public/purchase/info/journal.html?regNumber=%s", purNum)
		protocols = append(protocols, Puchase{pubDate: pubDate.String(), purNum: purNum, purName: purName, href: url, updDate: purchaseDate.String(), refused: true})
	}
	rows.Close()
	t.fixProtocols(protocols)
}
func (t *ParserEisFtp223) fixProtocols(protocols []Puchase) {
	for _, p := range protocols {
		t.writeAndSendPurchase(p)
	}
}

func (t *ParserEisFtp223) writeAndSendPurchase(p Puchase) {
	defer SaveStack()
	db, err := DbConnection()
	if err != nil {
		Logging(err)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT id FROM purchases WHERE purchase_number=$1", p.purNum)
	if err != nil {
		Logging(err)
		return
	}
	if rows.Next() {
		rows.Close()
		return
	}
	rows.Close()
	_, err = db.Exec("INSERT INTO purchases (id, purchase_number, date_update) VALUES (NULL, $1, $2)", p.purNum, p.updDate)
	if err != nil {
		Logging(err)
		return
	}
	t.addDoc++
	if p.refused {
		t.sendMessage(p)
	}
}

func (t *ParserEisFtp223) sendMessage(p Puchase) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		Logging(err)
	}
	msg := tgbotapi.NewMessage(ChannelId2, p.CreateMessage())
	msg.ParseMode = "html"
	_, err = bot.Send(msg)
	if err != nil {
		Logging(err)
	}
	t.sendDoc++
}
