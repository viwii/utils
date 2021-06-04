package utils

import (
	"math/rand"
)

// 生成区间为[start, end)
func GetRand(start int, end int) int {
	if (end - start) <= 0 {
		return start
	}

	return rand.Intn(end-start) + start
}
