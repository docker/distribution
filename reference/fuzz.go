package reference

func FuzzParseNormalizedNamed(data []byte) int {
	_, _ = ParseNormalizedNamed(string(data))
	return 1
}
