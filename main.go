package main

import "fmt"

func init() {
	CreateEnv()
}

func main() {
	defer SaveStack()
	Logging("Start work")
	p := ParserEisNew{}
	p.run()
	Logging(fmt.Sprintf("Add purchases %d", p.addDoc))
	Logging(fmt.Sprintf("Send purchases %d", p.sendDoc))
	Logging("End work")
}
