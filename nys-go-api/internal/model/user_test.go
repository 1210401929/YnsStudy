package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserJSONMatchesJavaBeanContract(t *testing.T) {
	user := User{GUID: "g1", Code: "demo", Name: "演示", UserNum: 7}
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{`"guid":"g1"`, `"code":"demo"`, `"name":"演示"`, `"usernum":7`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("用户 JSON 缺少 %s: %s", expected, text)
		}
	}
	if strings.Contains(text, `"password"`) {
		t.Fatalf("空密码字段不应输出: %s", text)
	}
}
