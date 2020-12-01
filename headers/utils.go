package headers

import (
	"strings"
	"unicode"
)

func makeHeaderID(name string) []byte {
	if name == "" {
		return nil
	}

	return []byte(strings.ToLower(strings.TrimSpace(name)))
}

// Values splits comma-delimited list of values and returns it as a
// list.
//
// Some headers allow merged values. For example, 'Accept-Encoding:
// deflate, gzip, br' is actually has 3 values: deflate, gzip and br.
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
