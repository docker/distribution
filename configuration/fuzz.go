package configuration

import (
	"bytes"
)

func ParserFuzzer(data []byte) int {
	rd := bytes.NewReader(data)
	_, _ = Parse(rd)
	return 1
}
