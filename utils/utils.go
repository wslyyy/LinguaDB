package utils

import (
	"math/rand"
	"time"
	"github.com/google/uuid"
)

// 通过传入的长度，生成100内的int类型随机数组
func GenRandEmbedding(length int) []float32 {
	nums := make([]float32, length)
	rand.Seed(time.Now().UnixNano())
	for i:=0;i<length;i++ {
		nums[i] = rand.Float32()
	}
	return nums
}

func GenUUID() string {
	uuid := uuid.New()
	key := uuid.String()
	return key
}
