package headers

import "strings"

func makeHeaderID(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func Values(value string) []string {
	values := strings.Split(value, ",")

	for k, v := range values {
		values[k] = strings.TrimSpace(v)
	}

	if len(values) == 0 || (len(values) == 1 && values[0] == "") {
		return nil
	}

	return values
}
