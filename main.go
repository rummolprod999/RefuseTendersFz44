package main

import "fmt"

func init() {
	CreateEnv()
}

func main() {
	defer SaveStack()
	Logging("Start work")
	m := ParserEisFtp{}
	m.run()
	Logging(fmt.Sprintf("Send purchases from ftp %d", m.sendDoc))
	p := ParserEisNew{}
	p.run()
	Logging(fmt.Sprintf("Add purchases from site %d", p.addDoc))
	Logging(fmt.Sprintf("Send purchases from site %d", p.sendDoc))

	Logging("End work")
}
