package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/security"
)

func TestCryptoDecryptsRequestAndEncryptsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cipher, err := security.NewAESCipher("7546455674406856", "7513956994549178")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"name":"噜噜"}`)
	encrypted, err := cipher.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"cipherText": encrypted})

	router := gin.New()
	router.Use(Crypto(cipher))
	router.POST("/echo", func(c *gin.Context) {
		var input map[string]string
		if err := c.ShouldBindJSON(&input); err != nil {
			t.Fatal(err)
		}
		c.JSON(http.StatusCreated, input)
	})

	request := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d", recorder.Code)
	}
	decrypted, err := cipher.Decrypt(recorder.Body.String())
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != `{"name":"噜噜"}` {
		t.Fatalf("response = %s", decrypted)
	}
}
