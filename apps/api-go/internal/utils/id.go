package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func NewID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewString())
}

func NewTimeBasedID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
