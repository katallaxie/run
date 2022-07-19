package utils

import "fmt"

// Strings ...
func Strings(m map[string]string) []string {
	ss := make([]string, len(m))
	for k, v := range m {
		ss = append(ss, fmt.Sprintf("%s=%s", k, v))
	}

	return ss
}
