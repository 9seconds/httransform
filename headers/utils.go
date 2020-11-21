package headers

import "strings"

func makeHeaderID(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
