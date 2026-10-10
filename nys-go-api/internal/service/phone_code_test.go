package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
)

func newPhoneCodeTestService(t *testing.T) (*Service, *int32) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var sent int32
	sms := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sent, 1)
	}))
	t.Cleanup(sms.Close)
	cfg := &config.Config{External: config.ExternalConfig{SMSURL: sms.URL, SMSCodeExpirationSeconds: 300}}
	return &Service{Config: cfg, Cache: cache.NewMemoryStore(), HTTP: sms.Client()}, &sent
}

func testContext(ip string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)
	c.Request.RemoteAddr = ip + ":1234"
	return c
}

func TestSendPhoneCodeCooldownAndIPLimit(t *testing.T) {
	service, sent := newPhoneCodeTestService(t)
	if result := service.SendPhoneCode(testContext("1.1.1.1"), "13800000000"); result.IsError {
		t.Fatalf("第一次发送应成功: %s", result.ErrMsg)
	}
	if result := service.SendPhoneCode(testContext("1.1.1.1"), "13800000000"); !result.IsError {
		t.Fatal("60 秒内重复发送应被拒绝")
	}
	// 同一 IP 换不同手机号，超过每小时上限后被拒绝
	for i := 1; i < phoneCodeIPLimit; i++ {
		phone := "1390000" + string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
		if result := service.SendPhoneCode(testContext("1.1.1.1"), phone); result.IsError {
			t.Fatalf("第 %d 次发送不应被拒绝: %s", i+1, result.ErrMsg)
		}
	}
	if result := service.SendPhoneCode(testContext("1.1.1.1"), "13700000000"); !result.IsError {
		t.Fatal("同一 IP 超过每小时上限应被拒绝")
	}
	if got := atomic.LoadInt32(sent); got != phoneCodeIPLimit {
		t.Fatalf("短信接口应只被调用 %d 次, 实际 %d", phoneCodeIPLimit, got)
	}
}

func TestSendPhoneCodeDailyLimit(t *testing.T) {
	service, _ := newPhoneCodeTestService(t)
	ctx := context.Background()
	for i := 0; i < phoneCodeDailyLimit; i++ {
		// 跳过 60 秒冷却，只验证每天的上限
		_ = service.Cache.Delete(ctx, "login:code:cooldown:13800000000")
		if result := service.SendPhoneCode(testContext("2.2.2."+string(rune('0'+i))), "13800000000"); result.IsError {
			t.Fatalf("第 %d 次发送不应被拒绝: %s", i+1, result.ErrMsg)
		}
	}
	_ = service.Cache.Delete(ctx, "login:code:cooldown:13800000000")
	if result := service.SendPhoneCode(testContext("3.3.3.3"), "13800000000"); !result.IsError {
		t.Fatal("同一手机号超过每天上限应被拒绝")
	}
}

func TestLoginByPhoneCodeInvalidatesAfterTooManyFailures(t *testing.T) {
	service, _ := newPhoneCodeTestService(t)
	ctx := context.Background()
	_ = service.Cache.Set(ctx, "login:code:13800000000", "123456", 0)
	for i := 1; i < phoneCodeMaxAttempts; i++ {
		if result := service.LoginByPhoneCode(testContext("1.1.1.1"), "13800000000", "000000"); !result.IsError {
			t.Fatal("错误验证码不应登录成功")
		}
	}
	result := service.LoginByPhoneCode(testContext("1.1.1.1"), "13800000000", "000000")
	if !result.IsError || result.ErrMsg != "验证码错误次数过多，请重新获取!" {
		t.Fatalf("第 %d 次输错应作废验证码, 实际: %s", phoneCodeMaxAttempts, result.ErrMsg)
	}
	// 验证码已作废，即使输入正确也不能再用
	if result := service.LoginByPhoneCode(testContext("1.1.1.1"), "13800000000", "123456"); !result.IsError {
		t.Fatal("作废后的验证码不应再能登录")
	}
}

func TestRegisterRejectsReservedAccountPrefix(t *testing.T) {
	service, _ := newPhoneCodeTestService(t)
	for _, code := range []string{"$userPhone13800000000", "$userQQabc"} {
		if result := service.Register(context.Background(), "x", code, "pw", "pw"); !result.IsError {
			t.Fatalf("不应允许注册保留账号 %s", code)
		}
	}
}

func TestClientIPPrefersRealIP(t *testing.T) {
	c := testContext("9.9.9.9")
	c.Request.Header.Set("X-Forwarded-For", "6.6.6.6, 10.0.0.1")
	c.Request.Header.Set("X-Real-IP", "10.0.0.1")
	if ip := ClientIP(c); ip != "10.0.0.1" {
		t.Fatalf("应使用 X-Real-IP, 实际 %s", ip)
	}
}

func TestLoginLockedAfterTooManyFailures(t *testing.T) {
	service, _ := newPhoneCodeTestService(t)
	ctx := context.Background()
	for i := 0; i < loginFailAccountMax; i++ {
		_, _ = service.Cache.Increment(ctx, "login:fail:user:yulei", loginFailWindow)
	}
	// 账号已锁定：即使密码正确也不再查库比对（这里没有数据库，能返回说明在查库前就拦下了）
	if result := service.Login(testContext("1.1.1.1"), "YuLei", "anything"); result.ErrMsg != loginFailAccountText {
		t.Fatalf("账号应被锁定, 实际: %s", result.ErrMsg)
	}
	for i := 0; i < loginFailIPMax; i++ {
		_, _ = service.Cache.Increment(ctx, "login:fail:ip:5.5.5.5", loginFailIPWindow)
	}
	if result := service.Login(testContext("5.5.5.5"), "other", "anything"); result.ErrMsg != loginFailIPText {
		t.Fatalf("IP 应被限制, 实际: %s", result.ErrMsg)
	}
	if result := service.Login(testContext("6.6.6.6"), "other", ""); !result.IsError {
		t.Fatal("空密码应被拒绝")
	}
}
