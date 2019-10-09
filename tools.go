package main

import (
	"regexp"
	"time"
)

func findFromRegExp(s string, t string) string {
	r := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if match != nil && len(match) > 1 {
		r = match[1]
	}
	return r
}
func getTimeMoscowLayout(st string, l string) time.Time {
	var p = time.Time{}
	location, _ := time.LoadLocation("Europe/Moscow")
	p, err := time.ParseInLocation(l, st, location)
	if err != nil {
		Logging(err)
		return time.Time{}
	}

	return p
}
