package hls

import "strings"

func parseTagAttributes(raw string) map[string]string {
	attrs := make(map[string]string)

	for _, part := range splitAttributes(raw) {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 || kv[0] == "" {
			continue
		}

		attrs[kv[0]] = strings.Trim(strings.TrimSpace(kv[1]), "\"")
	}

	return attrs
}

func splitAttributes(raw string) []string {
	parts := make([]string, 0)
	start := 0
	inQuotes := false

	for i, r := range raw {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				parts = append(parts, raw[start:i])
				start = i + 1
			}
		}
	}

	parts = append(parts, raw[start:])
	return parts
}

func tagAttributeData(line string) (string, bool) {
	i := strings.Index(line, ":")
	if i == -1 || i == len(line)-1 {
		return "", false
	}

	return line[i+1:], true
}
