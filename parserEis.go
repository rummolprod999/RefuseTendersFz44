package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type ParserEis struct {
	addDoc  int
	sendDoc int
}

func (t *ParserEis) run() {
	for i := 1; i <= CountPage; i++ {
		url := fmt.Sprintf("%s%d", StartUrl, i)
		t.parsingPage(url)
	}
}

func (t *ParserEis) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Got empty string", p)
	}
}

func (t *ParserEis) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.parametrs div.registerBox.registerBoxBank.margBtm20").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})

}

func (t *ParserEis) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("td.descriptTenderTd dl dt a").First().Text())
	purNum = strings.Replace(purNum, "№ ", "", -1)
	purName := strings.TrimSpace(p.Find("td.descriptTenderTd dl dd:nth-of-type(2)").First().Text())
	pubDate := strings.TrimSpace(p.Find("td.amountTenderTd ul li:nth-of-type(1)").First().Text())
	pubDate = strings.TrimSpace(strings.Replace(pubDate, "Размещено:", "", -1))
	updDate := strings.TrimSpace(p.Find("td.amountTenderTd ul li:nth-of-type(2)").First().Text())
	updDate = strings.TrimSpace(strings.Replace(updDate, "Обновлено:", "", -1))
	hrefT := p.Find("td.descriptTenderTd dl dt a")
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
func (t *ParserEis) getPurchasePage(p Puchase44) {
	r := DownloadPage(p.href)
	if r != "" {
		t.checkPurchase(r, p)
	} else {
		Logging("Got empty string", p.href)
	}
}

func (t *ParserEis) checkPurchase(s string, p Puchase44) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		Logging(err, p.href)
		return
	}
	eventText := strings.TrimSpace(doc.Find("#event tbody tr").First().Text())
	if strings.Contains(eventText, "участника") && strings.Contains(eventText, "Протокол") && strings.Contains(eventText, "уклонившимся") {
		p.refused = true
	}
	t.writeAndSendPurchase(p)
}

func (t *ParserEis) writeAndSendPurchase(p Puchase44) {
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

func (t *ParserEis) sendMessage(p Puchase44) {
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
