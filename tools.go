package main

import regexp "regexp"

func findFromRegExp(s string, t string) string {
	r := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if match != nil && len(match) > 1 {
		r = match[1]
	}
	return r
}
