package v2

func FuzzParseForwardedHeader(data []byte) int {
	_, _, _ = parseForwardedHeader(string(data))
	return 1
}
