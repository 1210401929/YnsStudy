package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func writeResult(c *gin.Context, result model.Result) {
	c.JSON(http.StatusOK, result)
}

func readBody(c *gin.Context) (map[string]any, error) {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	value := make(map[string]any)
	if err := decoder.Decode(&value); err != nil {
		if err == io.EOF {
			return value, nil
		}
		return nil, err
	}
	return value, nil
}

func requireBody(c *gin.Context) (map[string]any, bool) {
	body, err := readBody(c)
	if err != nil {
		writeResult(c, model.Failure("请求参数错误:"+err.Error()))
		return nil, false
	}
	return body, true
}

func nestedMap(body map[string]any, key string) map[string]any {
	value, _ := body[key].(map[string]any)
	if value == nil {
		value = make(map[string]any)
	}
	return value
}

func stringParam(body map[string]any, key string) string {
	value := body[key]
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func intParam(body map[string]any, key string, fallback int) int {
	value := body[key]
	switch typed := value.(type) {
	case json.Number:
		parsed, err := strconv.Atoi(typed.String())
		if err == nil {
			return parsed
		}
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func int64Param(body map[string]any, key string) (int64, error) {
	value := stringParam(body, key)
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("缺少参数 %s", key)
	}
	return strconv.ParseInt(value, 10, 64)
}

func anySlice(value any) []any {
	items, _ := value.([]any)
	return items
}

func stringSlice(value any) []string {
	items := anySlice(value)
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprint(item))
	}
	return result
}

func mapSlice(value any) []map[string]any {
	items := anySlice(value)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			result = append(result, row)
		}
	}
	return result
}

func nestedAnySlices(value any) [][]any {
	items := anySlice(value)
	result := make([][]any, 0, len(items))
	for _, item := range items {
		result = append(result, anySlice(item))
	}
	return result
}
