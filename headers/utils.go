package headers

import (
	"strings"
	"unicode"
)

func makeHeaderID(name string) []byte {
	return []byte(strings.ToLower(strings.TrimSpace(name)))
}

func Values(value string) []string {
	values := strings.Split(value, ",")

	for k, v := range values {
		values[k] = strings.TrimLeftFunc(v, unicode.IsSpace)
	}

	if len(values) == 0 || (len(values) == 1 && values[0] == "") {
		return nil
	}

	return values
}
