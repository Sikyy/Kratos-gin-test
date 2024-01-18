package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 随机数
// 生成min与max之间的整数（包含）
func GenRandomInt(min, max int) int {
	if min == max {
		return min
	}
	// 为了保险取两个值之间大的那个作为max
	randNum := rand.Intn(GetMaxInt(min, max)-min) + min
	return randNum
}

func GetMaxInt(min, max int) int {
	if max >= min {
		return max
	}
	return min
}

func GetMinInt(min, max int) int {
	if min <= max {
		return min
	}
	return max
}
