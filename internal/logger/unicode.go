package logger

import "encoding/json"

func decodeUnicode(s string) string {
	var decoded string
	if err := json.Unmarshal([]byte(`"`+s+`"`), &decoded); err != nil {
		return s
	}
	return decoded
}
