package main

import "fmt"

type Puchase struct {
	purNum  string
	purName string
	pubDate string
	updDate string
	href    string
	typeFz  int
	refused bool
}

func (p *Puchase) CheckPurchase() bool {
	if p.purNum == "" || p.purName == "" || p.pubDate == "" || p.href == "" {
		Logging(fmt.Sprintf("The purchase is bad %+v", p))
		return false
	}
	return true
}

func (p *Puchase) CreateMessage() string {
	return fmt.Sprintf("<b>Номер закупки</b>: %s\n<b>Наименование закупки</b>: %s\n<b>Дата публикации</b>: %s\n<b>Дата обновления</b>: %s\n<b>Ссылка</b>: %s", p.purNum, p.purName, p.pubDate, p.updDate, p.href)
}
