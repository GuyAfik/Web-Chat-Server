package utils

import (
	"math/rand"
	"time"
	"strings"
)

func GetRandom() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetRandomI64() int64 {
	return GetRandom().Int63()
}

func StringArrayToString(stringsSlice []string, delimiter string) string {
	return strings.Join(stringsSlice, delimiter)
}

func ParseMessageBody(message, delimiter string) []string {
	return strings.Split(message, delimiter)
}