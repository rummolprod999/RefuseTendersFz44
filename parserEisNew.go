package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

type ParserEisNew struct {
	addDoc  int
	sendDoc int
}

func (t *ParserEisNew) run() {
	for i := 1; i <= CountPage; i++ {
		url := fmt.Sprintf("%s%d", StartUrl, i)
		t.parsingPage(url)
	}
}

func (t *ParserEisNew) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Got empty string", p)
	}
}

func (t *ParserEisNew) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.search-registry-entry-block > div.row").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})

}

func (t *ParserEisNew) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("div.registry-entry__header-mid__number a").First().Text())
	purNum = strings.Replace(purNum, "№ ", "", -1)
	if len(purNum) < 18 {
		return
	}
	purName := strings.TrimSpace(p.Find("div:contains('Объект закупки') + div.registry-entry__body-value").First().Text())
	pubDate := strings.TrimSpace(p.Find("div.data-block > div:contains('Размещено') + div").First().Text())
	updDate := strings.TrimSpace(p.Find("div.data-block > div:contains('Обновлено') + div").First().Text())
	updDate = strings.TrimSpace(strings.Replace(updDate, "Обновлено:", "", -1))
	hrefT := p.Find("div.registry-entry__header-mid__number a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element have no href attribute", hrefT.Text())
		return
	}
	href = strings.Replace(href, "common-info", "event-journal", -1)
	href = fmt.Sprintf("http://zakupki.gov.ru%s", href)
	purch := Puchase44{href: href, pubDate: pubDate, updDate: updDate, purName: purName, purNum: purNum}
	if purch.CheckPurchase() {
		t.getPurchasePage(purch)
	}
}

func (t *ParserEisNew) getPurchasePage(p Puchase44) {
	r := DownloadPage(p.href)
	if r != "" {
		t.checkPurchase(r, p)
	} else {
		Logging("Got empty string", p.href)
	}
}

func (t *ParserEisNew) checkPurchase(s string, p Puchase44) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		Logging(err, p.href)
		return
	}
	timeNow := time.Now()
	ft := timeNow.Format("02.01.2006")
	doc.Find("div.tabBoxWrapper table.table tbody tr.table__row.table__row-body").Each(func(i int, s *goquery.Selection) {
		textEvent := s.Text()
		if strings.Contains(textEvent, "участника") && strings.Contains(textEvent, "Протокол") && strings.Contains(textEvent, "уклонившимся") && strings.Contains(textEvent, ft) {
			p.refused = true
		}

	})
	newP := strings.TrimSpace(doc.Find("div.tabBoxWrapper table.table tbody tr.table__row.table__row-body").First().Text())
	extractDate := findFromRegExp(newP, `(\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2})`)
	updDateString := ""
	if extractDate != "" {
		updDate := getTimeMoscowLayout(extractDate, "02.01.2006 15:04")
		offsetPur := findFromRegExp(newP, `МСК([+-]\d{1,2})`)
		var valOffset int64 = 0
		if offsetPur != "" {
			valOffset, _ = strconv.ParseInt(offsetPur, 10, 32)
			if valOffset != 0 {
				updDate = updDate.Add(time.Hour * time.Duration(valOffset*-1))
			}
		}
		updDateString = updDate.Format("02.01.2006 15:04")
	}
	if updDateString != "" {
		p.updDate = updDateString
	} else {
		p.updDate = newP
	}
	t.writeAndSendPurchase(p)
}

func (t *ParserEisNew) writeAndSendPurchase(p Puchase44) {
	db, err := DbConnection()
	if err != nil {
		Logging(err)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT id FROM purchases WHERE purchase_number=$1 AND date_update=$2", p.purNum, p.updDate)
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

func (t *ParserEisNew) sendMessage(p Puchase44) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		Logging(err)
	}
	msg := tgbotapi.NewMessage(ChannelId, p.CreateMessage())
	msg.ParseMode = "html"
	_, err = bot.Send(msg)
	if err != nil {
		Logging(err)
	}
	t.sendDoc++
}
