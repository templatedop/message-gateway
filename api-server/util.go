package router

import "strings"

func Sanitize(str string) string {
	san := strings.ReplaceAll(str, " ", "_")
	san = strings.ReplaceAll(san, "-", "_")

	return strings.ToLower(san)
}

func Split(str string) []string {
	return strings.Split(strings.ReplaceAll(str, " ", ""), ",")
}
