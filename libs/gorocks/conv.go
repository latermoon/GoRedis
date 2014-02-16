package gorocks

// #include "rocksdb/c.h"
import "C"

func boolToUchar(b bool) C.uchar {
	if b {
		return 1
	}
	return 0
}

func ucharToBool(uc C.uchar) bool {
	if uc == 0 {
		return false
	}
	return true
}

func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
