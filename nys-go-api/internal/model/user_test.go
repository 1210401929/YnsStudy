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

func TestPublicDisplayNameMasksPhoneNumbers(t *testing.T) {
	cases := map[string]string{
		"15611122742":   "手机用户2742",
		" 18531349133 ": "手机用户9133",
		"余南松":           "余南松",
		"1561112274":    "1561112274",   // 只有 10 位，不是手机号
		"156111227421":  "156111227421", // 12 位
		"手机用户2742":      "手机用户2742",
	}
	for input, want := range cases {
		if got := PublicDisplayName(input); got != want {
			t.Errorf("PublicDisplayName(%q) = %q, want %q", input, got, want)
		}
	}
}
