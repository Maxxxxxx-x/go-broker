package transport

func TrimLine(str string) string {
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '\n' || str[i] == '\r' {
			str = str[:i]
		} else {
			break
		}
	}

	return str
}

func ExtractLine(buf *[]byte) (string, bool) {
	for i := range len(*buf) {
		if (*buf)[i] == '\n' {
			line := string((*buf)[:i])
			*buf = (*buf)[i+1:]
			return line, true
		}
	}
	return "", false
}
