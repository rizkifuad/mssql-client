package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

func ParseInterface(arg *interface{}) string {
	var result string
	switch v := (*arg).(type) {
	case nil:
		result = "NULL"
		break
	case bool:
		if v {
			result = "1"
		} else {
			result = "0"
		}
		break
	case []byte:
		result = fmt.Sprintf("%v", string(v))
		break
	case time.Time:
		result = fmt.Sprintf("%v", v.Format("2006-01-02 15:04:05.999"))
		break
	case bytes.Buffer:
		result = fmt.Sprintf("%v", v)
		break
	default:
		result = fmt.Sprintf("%+v", v)
	}
	return result
}

func ParseValue(args ...*interface{}) string {
	a := args[0]
	return ParseInterface(a)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
