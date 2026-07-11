package cbridge

const Backend = "go"

func FNV1a(data []byte) uint32 {
	var hash uint32 = 2166136261
	for _, b := range data {
		hash ^= uint32(b)
		hash *= 16777619
	}
	return hash
}
