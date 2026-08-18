package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Lookup(row map[string]any, key string) any {
	if value, ok := row[key]; ok {
		return value
	}
	for candidate, value := range row {
		if strings.EqualFold(candidate, key) {
			return value
		}
	}
	return nil
}

func StringValue(row map[string]any, key string) string {
	value := Lookup(row, key)
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(typed)
	}
}

func IntValue(row map[string]any, key string) int {
	return int(Int64Value(row, key))
}

func Int64Value(row map[string]any, key string) int64 {
	value := Lookup(row, key)
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		result, _ := typed.Int64()
		return result
	case []byte:
		result, _ := strconv.ParseInt(string(typed), 10, 64)
		return result
	case string:
		result, _ := strconv.ParseInt(typed, 10, 64)
		return result
	default:
		result, _ := strconv.ParseInt(fmt.Sprint(typed), 10, 64)
		return result
	}
}
