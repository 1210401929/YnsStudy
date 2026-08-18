package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/security"
)

const encryptedField = "cipherText"

func Crypto(aesCipher *security.AESCipher) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || strings.HasPrefix(c.Request.URL.Path, "/uploadFile/") || strings.HasPrefix(c.ContentType(), "multipart/") {
			c.Next()
			return
		}
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "读取请求内容失败", "data": nil})
			return
		}
		c.Request.Body.Close()

		var envelope map[string]json.RawMessage
		var cipherText string
		if len(raw) > 0 && json.Unmarshal(raw, &envelope) == nil {
			if value, ok := envelope[encryptedField]; ok {
				_ = json.Unmarshal(value, &cipherText)
			}
		}
		if cipherText == "" {
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			c.Next()
			return
		}

		plain, err := aesCipher.Decrypt(cipherText)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求数据解密失败", "data": nil})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Request.ContentLength = int64(len(plain))
		c.Request.Header.Set("Content-Type", "application/json")

		original := c.Writer
		buffered := &encryptedResponseWriter{ResponseWriter: original, status: http.StatusOK}
		c.Writer = buffered
		c.Next()

		encrypted, err := aesCipher.Encrypt(buffered.body.Bytes())
		if err != nil {
			original.Header().Set("Content-Type", "application/json; charset=utf-8")
			original.WriteHeader(http.StatusInternalServerError)
			_, _ = original.Write([]byte(`{"code":500,"message":"响应数据加密失败","data":null}`))
			return
		}
		original.Header().Del("Content-Length")
		original.Header().Set("Content-Type", "text/plain; charset=utf-8")
		original.WriteHeader(buffered.status)
		_, _ = original.Write([]byte(encrypted))
	}
}

type encryptedResponseWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
	size   int
	wrote  bool
}

func (w *encryptedResponseWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	w.wrote = true
}

func (w *encryptedResponseWriter) WriteHeaderNow() {
	if !w.wrote {
		w.WriteHeader(w.status)
	}
}

func (w *encryptedResponseWriter) Write(data []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}

func (w *encryptedResponseWriter) WriteString(value string) (int, error) {
	return w.Write([]byte(value))
}

func (w *encryptedResponseWriter) Status() int { return w.status }
func (w *encryptedResponseWriter) Size() int   { return w.size }
func (w *encryptedResponseWriter) Written() bool {
	return w.wrote
}
