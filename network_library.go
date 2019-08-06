package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func DownloadPage(url string) string {
	count := 0
	var st string
	for {
		//fmt.Println("Start download file")
		if count > 5 {
			Logging(fmt.Sprintf("The file was not downloaded in %d attemps %s", count, url))
			return st
		}
		st = GetPageUA(url)
		if st == "" {
			count++
			Logging("Got empty page", url)
			time.Sleep(time.Second * 5)
			continue
		}
		return st

	}
	return st
}

func GetPageUA(url string) (ret string) {
	defer func() {
		if r := recover(); r != nil {
			Logging(fmt.Sprintf("Was panic, recovered value: %v", r))
			ret = ""
		}
	}()
	var st string
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logging("Error request", url, err)
		return st
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko)")
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		Logging("Error download", url, err)
		return st
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logging("Error reading", url, err)
		return st
	}

	return string(body)
}
