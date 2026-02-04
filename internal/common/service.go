package common

import (
	"time"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateTimestamp() time.Time {
	return time.Now()
}
