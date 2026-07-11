package cbridge

/*
#include <stdint.h>

static uint32_t fnv1a(const unsigned char *data, int len) {
    uint32_t hash = 2166136261u;
    for (int i = 0; i < len; i++) {
        hash ^= data[i];
        hash *= 16777619u;
    }
    return hash;
}
*/
import "C"
import "unsafe"

const Backend = "cgo"

func FNV1a(data []byte) uint32 {
	if len(data) == 0 {
		return 2166136261
	}
	ptr := (*C.uchar)(unsafe.Pointer(&data[0]))
	return uint32(C.fnv1a(ptr, C.int(len(data))))
}
