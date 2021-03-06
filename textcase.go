package flagtags

import (
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toKebabCase(str string) string {
	return strings.ReplaceAll(toSnakeCase(str), "_", "-")
}

func toScreamingSnakeCase(str string) string {
	return strings.ToUpper(toSnakeCase(str))
}
