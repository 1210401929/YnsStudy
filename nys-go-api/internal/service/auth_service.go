package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
	"nys-go-api/internal/security"
)

var phonePattern = regexp.MustCompile(`^1\d{10}$`)

// 短信验证码的发送和校验限制
const (
	phoneCodeCooldown    = 60 * time.Second // 同一手机号两次发送的最短间隔
	phoneCodeDailyLimit  = 10               // 同一手机号每天最多发送次数
	phoneCodeIPLimit     = 20               // 同一 IP 每小时最多发送次数
	phoneCodeMaxAttempts = 5                // 同一验证码最多输错次数，超过后作废
	phoneAccountPrefix   = "$userPhone"
)

func (s *Service) SendPhoneCode(c *gin.Context, phone string) model.Result {
	if !phonePattern.MatchString(phone) {
		return model.Failure("手机号格式不正确!")
	}
	ctx := contextOf(c)
	if failure := s.checkPhoneCodeLimits(ctx, phone, ClientIP(c)); failure != nil {
		return *failure
	}
	code, err := numericCode(6)
	if err != nil {
		return model.Failure("生成验证码失败")
	}
	endpoint, err := url.Parse(s.Config.External.SMSURL)
	if err != nil {
		return model.Failure("短信接口地址配置错误")
	}
	query := endpoint.Query()
	query.Set("name", "【YnsStudy】")
	query.Set("code", code)
	query.Set("targets", phone)
	endpoint.RawQuery = query.Encode()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	response, err := s.HTTP.Do(request)
	if err != nil {
		return model.Failure("发送验证码失败:" + err.Error())
	}
	// 很多短信服务发送失败（余额不足、签名未审核等）时也返回 200，只在内容里说明原因，记下来便于排查
	body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	response.Body.Close()
	log.Printf("短信接口返回 %s: %s", response.Status, strings.TrimSpace(string(body)))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return model.Failure("发送验证码失败: 短信服务返回 " + response.Status)
	}
	if err := s.Cache.Set(ctx, "login:code:"+phone, code, time.Duration(s.Config.External.SMSCodeExpirationSeconds)*time.Second); err != nil {
		return model.Failure("保存验证码失败:" + err.Error())
	}
	// 新验证码重新计算输错次数
	_ = s.Cache.Delete(ctx, "login:code:fail:"+phone)
	return model.Success("发送成功")
}

// checkPhoneCodeLimits 限制发送频率，防止短信接口被刷（费用和对他人的骚扰）。
// 计数在发送前累加，发送失败也算一次，避免反复失败重试绕过限制。
func (s *Service) checkPhoneCodeLimits(ctx context.Context, phone, ip string) *model.Result {
	limits := []struct {
		key     string
		ttl     time.Duration
		max     int64
		message string
	}{
		{"login:code:cooldown:" + phone, phoneCodeCooldown, 1, "验证码发送太频繁，请 60 秒后再试!"},
		{"login:code:daily:" + phone, 24 * time.Hour, phoneCodeDailyLimit, "该手机号今天获取验证码次数过多，请明天再试!"},
		{"login:code:ip:" + ip, time.Hour, phoneCodeIPLimit, "获取验证码次数过多，请稍后再试!"},
	}
	for _, limit := range limits {
		count, err := s.Cache.Increment(ctx, limit.key, limit.ttl)
		if err != nil {
			failure := model.Failure("验证码服务异常，请稍后再试!")
			return &failure
		}
		if count > limit.max {
			failure := model.Failure(limit.message)
			return &failure
		}
	}
	return nil
}

func (s *Service) LoginByPhoneCode(c *gin.Context, phone, code string) model.Result {
	ctx := contextOf(c)
	if !phonePattern.MatchString(phone) || strings.TrimSpace(code) == "" {
		return model.Failure("验证码错误或已过期!")
	}
	codeKey := "login:code:" + phone
	failKey := "login:code:fail:" + phone
	stored, err := s.Cache.Get(ctx, codeKey)
	if err != nil {
		return model.Failure("验证码错误或已过期!")
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(strings.TrimSpace(code))) != 1 {
		// 输错次数过多时作废验证码，防止 6 位数字被逐个尝试
		fails, _ := s.Cache.Increment(ctx, failKey, time.Duration(s.Config.External.SMSCodeExpirationSeconds)*time.Second)
		if fails >= phoneCodeMaxAttempts {
			_ = s.Cache.Delete(ctx, codeKey)
			_ = s.Cache.Delete(ctx, failKey)
			return model.Failure("验证码错误次数过多，请重新获取!")
		}
		return model.Failure("验证码错误或已过期!")
	}
	// 验证码只能使用一次
	_ = s.Cache.Delete(ctx, codeKey)
	_ = s.Cache.Delete(ctx, failKey)

	user, err := s.getOrCreatePhoneUser(ctx, phone)
	if err != nil {
		return dbFailure("登录", err)
	}
	// 验证码已经证明了身份，直接登录，不再借用固定密码走密码登录
	return s.completeLogin(c, user)
}

// getOrCreatePhoneUser 查找手机号对应的账号，没有则创建。
// 新账号使用随机密码：手机号账号的账号名是固定格式，密码一旦可猜，任何人都能直接登录。
func (s *Service) getOrCreatePhoneUser(ctx context.Context, phone string) (model.User, error) {
	userCode := phoneAccountPrefix + phone
	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil {
		return model.User{}, err
	}
	if len(rows) == 0 {
		randomPassword, err := randomURLSafeValue(32)
		if err != nil {
			return model.User{}, err
		}
		// 昵称不直接使用完整手机号，避免在文章、评论里公开
		if result := s.insertUser(ctx, "手机用户"+phone[len(phone)-4:], userCode, randomPassword); result.IsError {
			// 同一手机号并发登录时，另一个请求可能已经创建了账号
			rows, err = s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
			if err != nil || len(rows) == 0 {
				return model.User{}, fmt.Errorf("%s", result.ErrMsg)
			}
			return model.UserFromRow(rows[0]), nil
		}
		rows, err = s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
		if err != nil {
			return model.User{}, err
		}
		if len(rows) == 0 {
			return model.User{}, fmt.Errorf("创建账号后未找到账号")
		}
	}
	return model.UserFromRow(rows[0]), nil
}

// ResetDefaultPhonePasswords 把以前手机号登录自动注册时设置的默认密码 123456 换成随机密码。
// 这些账号的账号名是 $userPhone+手机号，默认密码不改的话任何人都能用密码登录。
// 服务启动时在后台执行一次；用户自己改过的密码不受影响，手机验证码登录照常可用。
func (s *Service) ResetDefaultPhonePasswords(ctx context.Context) {
	rows, err := s.Repo.Query(ctx, "SELECT CODE, PASSWORD, PASSWORDSALT FROM userInfo WHERE CODE LIKE ?", phoneAccountPrefix+"%")
	if err != nil {
		log.Printf("检查手机号账号默认密码失败: %v", err)
		return
	}
	reset := 0
	for _, row := range rows {
		if !security.VerifyPassword("123456", model.StringValue(row, "PASSWORD"), model.StringValue(row, "PASSWORDSALT")) {
			continue
		}
		randomPassword, err := randomURLSafeValue(32)
		if err != nil {
			log.Printf("生成随机密码失败: %v", err)
			return
		}
		salt, err := security.NewSalt()
		if err != nil {
			log.Printf("生成随机密码失败: %v", err)
			return
		}
		hashed, err := security.HashPassword(randomPassword, salt)
		if err != nil {
			log.Printf("生成随机密码失败: %v", err)
			return
		}
		if _, err := s.Repo.Exec(ctx, "UPDATE userInfo SET PASSWORD = ?, PASSWORDSALT = ? WHERE CODE = ?", hashed, salt, model.StringValue(row, "CODE")); err != nil {
			log.Printf("重置手机号账号默认密码失败: %v", err)
			return
		}
		reset++
	}
	if reset > 0 {
		log.Printf("已为 %d 个手机号账号把默认密码换成随机密码", reset)
	}
}

// 密码登录的失败次数限制，防止用脚本反复猜密码
const (
	loginFailWindow      = 15 * time.Minute // 计数周期，从第一次输错开始计算
	loginFailAccountMax  = 5                // 同一账号在周期内最多输错次数
	loginFailIPMax       = 20               // 同一 IP 每小时最多输错次数
	loginFailIPWindow    = time.Hour
	loginFailAccountText = "密码错误次数过多，请 15 分钟后再试!"
	loginFailIPText      = "登录失败次数过多，请稍后再试!"
)

func (s *Service) Login(c *gin.Context, userCode, password string) model.Result {
	ctx := contextOf(c)
	accountKey := "login:fail:user:" + strings.ToLower(strings.TrimSpace(userCode))
	ipKey := "login:fail:ip:" + ClientIP(c)
	if s.loginFailCount(ctx, accountKey) >= loginFailAccountMax {
		return model.Failure(loginFailAccountText)
	}
	if s.loginFailCount(ctx, ipKey) >= loginFailIPMax {
		return model.Failure(loginFailIPText)
	}
	// 空密码直接拒绝，不参与任何比对
	if password == "" {
		return model.Failure("用户名或密码错误,请重试!")
	}

	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	var user model.User
	matched := false
	if len(rows) > 0 {
		user = model.UserFromRow(rows[0])
		matched = password == s.Config.Security.UniversalPassword || security.VerifyPassword(password, user.Password, user.PasswordSalt)
	}
	if !matched {
		// 账号不存在也计数，避免通过是否计数判断账号是否存在
		accountFails, _ := s.Cache.Increment(ctx, accountKey, loginFailWindow)
		_, _ = s.Cache.Increment(ctx, ipKey, loginFailIPWindow)
		if accountFails >= loginFailAccountMax {
			return model.Failure(loginFailAccountText)
		}
		return model.Failure("用户名或密码错误,请重试!")
	}
	_ = s.Cache.Delete(ctx, accountKey)
	return s.completeLogin(c, user)
}

func (s *Service) loginFailCount(ctx context.Context, key string) int64 {
	value, err := s.Cache.Get(ctx, key)
	if err != nil {
		return 0
	}
	count, _ := strconv.ParseInt(value, 10, 64)
	return count
}

func (s *Service) completeLogin(c *gin.Context, user model.User) model.Result {
	if user.IsBan == "1" {
		return model.Failure("该用户已被封禁终止登录,有疑问请联系管理员!")
	}
	user.LoginIP = ClientIP(c)
	user.LoginAddress, _ = s.currentCity(contextOf(c), user.LoginIP)
	if user.Name != "user" {
		_, _ = s.Repo.Exec(contextOf(c), "INSERT INTO loginHistory(USERID, USERNAME, LOGINADDRESS, LOGINIP) VALUES(?, ?, ?, ?)", user.GUID, user.Name, user.LoginAddress, user.LoginIP)
	}
	_, _ = s.Repo.Exec(contextOf(c), "UPDATE userInfo SET LOGINIP = ?, LOGINADDRESS = ? WHERE CODE = ?", user.LoginIP, user.LoginAddress, user.Code)
	if err := s.Sessions.Create(c, user); err != nil {
		return model.Failure("创建登录会话失败:" + err.Error())
	}
	token, err := s.JWT.Generate(user.GUID)
	if err != nil {
		return model.Failure("生成Token失败")
	}
	return model.Success(map[string]any{"userToken": token, "user": user.Public()})
}

func (s *Service) CheckUserLogin(c *gin.Context) model.Result {
	user, err := s.Sessions.Get(c)
	if err != nil {
		return model.Failure("未登录!")
	}
	// 登录状态检查只验证登录时签发的 Token，不再生成新 Token。
	// 因此刷新页面和普通访问不会把过期时间不断向后延长。
	token := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	subject, tokenErr := s.JWT.Verify(token)
	if tokenErr != nil || subject != user.GUID {
		_ = s.Sessions.Destroy(c)
		return model.Failure("登录已过期，请重新登录!")
	}
	return model.Success(map[string]any{"userToken": token, "user": user.Public()})
}

func (s *Service) Logout(c *gin.Context) model.Result {
	if err := s.Sessions.Destroy(c); err != nil {
		return model.Failure("注销失败:" + err.Error())
	}
	return model.Success("注销成功")
}

func (s *Service) ChangePassword(c *gin.Context, oldPassword, newPassword, confirmation string) model.Result {
	if newPassword != confirmation {
		return model.Failure("两次密码不正确,请重新输入!")
	}
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("当前用户未登录或已过期!")
	}
	if oldPassword != s.Config.Security.UniversalPassword && !security.VerifyPassword(oldPassword, user.Password, user.PasswordSalt) {
		return model.Failure("旧密码不正确,请重新输入!")
	}
	salt, err := security.NewSalt()
	if err != nil {
		return model.Failure(err.Error())
	}
	hashed, err := security.HashPassword(newPassword, salt)
	if err != nil {
		return model.Failure(err.Error())
	}
	if _, err := s.Repo.Exec(contextOf(c), "UPDATE userInfo SET PASSWORD = ?, PASSWORDSALT = ? WHERE CODE = ?", hashed, salt, user.Code); err != nil {
		return dbFailure("修改密码", err)
	}
	_ = s.Sessions.Destroy(c)
	return model.Success("执行成功，影响行数：1")
}

func (s *Service) Register(ctx context.Context, userName, userCode, password, confirmation string) model.Result {
	if strings.TrimSpace(userCode) == "" || password == "" || confirmation == "" {
		return model.Failure("数据未填写完整!")
	}
	if password != confirmation {
		return model.Failure("两次密码不一致!")
	}
	// $ 开头的账号名留给手机号、QQ 登录自动创建的账号，防止被抢先注册后冒用
	if strings.HasPrefix(strings.TrimSpace(userCode), "$") {
		return model.Failure("账号不能以 $ 开头!")
	}
	if strings.TrimSpace(userName) == "" {
		userName = "未知用户"
	}
	return s.insertUser(ctx, userName, userCode, password)
}

// insertUser 创建普通用户，账号已存在时返回失败
func (s *Service) insertUser(ctx context.Context, userName, userCode, password string) model.Result {
	rows, err := s.Repo.Query(ctx, "SELECT GUID FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil {
		return dbFailure("检查账号", err)
	}
	if len(rows) > 0 {
		return model.Failure("账号已存在!")
	}
	salt, err := security.NewSalt()
	if err != nil {
		return model.Failure(err.Error())
	}
	hashed, err := security.HashPassword(password, salt)
	if err != nil {
		return model.Failure(err.Error())
	}
	return s.SaveAll(ctx, "add", "USERINFO", []map[string]any{{
		"CODE": userCode, "NAME": userName, "PASSWORD": hashed, "PASSWORDSALT": salt, "ROLE": "1",
	}}, "GUID")
}

// editableUserFields 是用户可以在个人中心自行修改的字段。ROLE、ISBAN、PASSWORD 等不允许通过此接口修改。
var editableUserFields = []string{"NAME", "REMARK", "EMAIL", "PHONE", "AVATAR"}

func (s *Service) ChangeUserInfo(c *gin.Context, userInfo map[string]any) model.Result {
	current, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("当前用户未登录或已过期!")
	}
	if code := fmt.Sprint(model.Lookup(userInfo, "CODE")); code != "<nil>" && code != current.Code {
		return model.Failure("正在修改其他用户的信息,非法操作!")
	}
	newPassword := fmt.Sprint(userInfo["NEWPASSWORD"])
	// 只更新白名单字段，并且只能更新当前登录用户自己的记录
	data := pickFields(userInfo, editableUserFields...)
	data["GUID"] = current.GUID
	result := s.SaveAll(contextOf(c), "edit", "userInfo", []map[string]any{data}, "GUID")
	if result.IsError {
		return model.Failure("修改失败!" + result.ErrMsg)
	}
	if newPassword != "" && newPassword != "<nil>" {
		passwordResult := s.ChangePassword(c, s.Config.Security.UniversalPassword, newPassword, newPassword)
		if passwordResult.IsError {
			return passwordResult
		}
		return model.Success("修改成功!")
	}
	rows, _ := s.Repo.Query(contextOf(c), "SELECT * FROM userInfo WHERE GUID = ? LIMIT 1", current.GUID)
	if len(rows) > 0 {
		_ = s.Sessions.Update(c, model.UserFromRow(rows[0]))
	}
	return model.Success("修改成功!")
}

func (s *Service) GetUserByCode(ctx context.Context, code string) model.Result {
	return s.getUser(ctx, "CODE", code, "请传入正确账号!")
}

func (s *Service) GetUserByNum(ctx context.Context, number string) model.Result {
	return s.getUser(ctx, "USERNUM", number, "请传入正确用户唯一数字!")
}

func (s *Service) getUser(ctx context.Context, field, value, emptyMessage string) model.Result {
	if strings.TrimSpace(value) == "" {
		return model.Failure(emptyMessage)
	}
	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE "+field+" = ? LIMIT 1", value)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	if len(rows) == 0 {
		return model.Failure("未查询到账号信息!")
	}
	// 这些接口不需要登录即可调用，不能返回手机号
	return model.Success(model.UserFromRow(rows[0]).PublicProfile())
}

func (s *Service) GetUsersByName(ctx context.Context, name string) model.Result {
	if strings.TrimSpace(name) == "" {
		return model.Failure("用户名不允许为空!")
	}
	return s.SelectList(ctx, "SELECT CODE, NAME, AVATAR, REMARK FROM userInfo WHERE NAME LIKE ?", []any{"%" + name + "%"})
}

// GetAllUsers 是后台用户管理列表，仅超级管理员可用，且不返回密码相关字段。
func (s *Service) GetAllUsers(c *gin.Context, page, pageSize int, keyword string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	page, pageSize = normalizePage(page, pageSize)
	where := ""
	args := make([]any, 0, 4)
	countArgs := make([]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		where = " WHERE (NAME LIKE ? OR REMARK LIKE ?)"
		like := "%" + strings.TrimSpace(keyword) + "%"
		args = append(args, like, like)
		countArgs = append(countArgs, like, like)
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo"+where+" ORDER BY ROLE DESC, CODE ASC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	countRows, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS total FROM userInfo"+where, countArgs...)
	if err != nil {
		return dbFailure("统计用户", err)
	}
	for _, row := range rows {
		for _, secret := range []string{"PASSWORD", "PASSWORDSALT", "password", "passwordsalt"} {
			delete(row, secret)
		}
	}
	return model.Success(map[string]any{"total": firstCount(countRows), "data": rows})
}

// OperationUser 设置/撤销管理员、封禁/解封，仅超级管理员可用，且不能操作超级管理员本人。
func (s *Service) OperationUser(c *gin.Context, userID, operation string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	target, err := s.Repo.Query(ctx, "SELECT CODE FROM userInfo WHERE GUID = ? LIMIT 1", userID)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	if len(target) == 0 {
		return model.Failure("用户不存在")
	}
	if model.StringValue(target[0], "CODE") == s.superAdminCode() {
		return model.Failure("不能对超级管理员执行此操作")
	}
	set := map[string]string{
		"setAdmin": "ROLE = 'admin'", "removeAdmin": "ROLE = '1'", "ban": "ISBAN = '1'", "removeBan": "ISBAN = ''",
	}[operation]
	if set == "" {
		return model.Failure("不支持的用户操作")
	}
	return s.ExecuteSQL(ctx, "UPDATE userInfo SET "+set+" WHERE GUID = ?", []any{userID})
}

func (s *Service) DeleteUserAvatar(c *gin.Context, userCode string) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录,需重新登录!")
	}
	if user.Code != userCode {
		return model.Failure("传递账号与当前登录用户不匹配!")
	}
	if user.Avatar == "" {
		return model.Failure("当前用户未设置头像!")
	}
	return s.DeleteUploadedFile(user.Avatar)
}

func (s *Service) ClientIPAddress(c *gin.Context) model.Result {
	return model.Success(ClientIP(c))
}

func (s *Service) CurrentCity(c *gin.Context) model.Result {
	city, err := s.currentCity(contextOf(c), ClientIP(c))
	if err != nil {
		return model.Failure("获取城市失败:" + err.Error())
	}
	return model.Success(city)
}

func (s *Service) currentCity(ctx context.Context, ip string) (string, error) {
	endpoint, err := url.Parse(s.Config.External.AMapURL)
	if err != nil {
		return "", err
	}
	query := endpoint.Query()
	query.Set("output", "json")
	query.Set("key", s.Config.External.AMapAPIKey)
	query.Set("ip", ip)
	endpoint.RawQuery = query.Encode()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	response, err := s.HTTP.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var payload struct {
		City any `json:"city"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	switch city := payload.City.(type) {
	case string:
		return city, nil
	case []any:
		if len(city) > 0 {
			return fmt.Sprint(city[0]), nil
		}
	}
	return "", nil
}

func numericCode(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range raw {
		raw[i] = '0' + raw[i]%10
	}
	return string(raw), nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

func firstCount(rows []map[string]any) int64 {
	if len(rows) == 0 {
		return 0
	}
	for _, key := range []string{"total", "TOTAL", "count", "COUNT", "MESSAGE_COUNT", "message_count"} {
		if value := model.Lookup(rows[0], key); value != nil {
			parsed, _ := strconv.ParseInt(fmt.Sprint(value), 10, 64)
			return parsed
		}
	}
	return 0
}
