package models

import (
	"crypto/rand"
	"fmt"
	"time"
)

// generateID 生成唯一标识符
func generateID() string {
	timestamp := time.Now().UnixNano()
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%d-%x", timestamp, bytes)
}
