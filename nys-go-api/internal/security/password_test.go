package security

import "testing"

func TestPasswordHashCompatibility(t *testing.T) {
	const salt = "AAECAwQFBgcICQoLDA0ODw=="
	hashed, err := HashPassword("123456", salt)
	if err != nil {
		t.Fatal(err)
	}
	if hashed != "mckHc+qc5mo1W9aipuVRIsPK9aFkeFyQsdWMLWO7hoA=" {
		t.Fatalf("散列结果发生变化: %s", hashed)
	}
	if !VerifyPassword("123456", hashed, salt) || VerifyPassword("wrong", hashed, salt) {
		t.Fatal("密码校验结果错误")
	}
}
