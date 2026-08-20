package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
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

func (s *Service) SendPhoneCode(ctx context.Context, phone string) model.Result {
	if !phonePattern.MatchString(phone) {
		return model.Failure("手机号格式不正确!")
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
	_, _ = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return model.Failure("发送验证码失败: 短信服务返回 " + response.Status)
	}
	if err := s.Cache.Set(ctx, "login:code:"+phone, code, time.Duration(s.Config.External.SMSCodeExpirationSeconds)*time.Second); err != nil {
		return model.Failure("保存验证码失败:" + err.Error())
	}
	return model.Success("发送成功")
}

func (s *Service) LoginByPhoneCode(c *gin.Context, phone, code string) model.Result {
	stored, err := s.Cache.Get(contextOf(c), "login:code:"+phone)
	if err != nil || !strings.EqualFold(stored, code) {
		return model.Failure("验证码错误或已过期!")
	}
	userCode := "$userPhone" + phone
	rows, err := s.Repo.Query(contextOf(c), "SELECT * FROM userInfo WHERE CODE = ?", userCode)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	if len(rows) == 0 {
		registered := s.Register(contextOf(c), phone, userCode, "123456", "123456")
		if registered.IsError {
			return registered
		}
	}
	result := s.Login(c, userCode, "123456")
	if !result.IsError {
		_ = s.Cache.Delete(contextOf(c), "login:code:"+phone)
	}
	return result
}

func (s *Service) Login(c *gin.Context, userCode, password string) model.Result {
	rows, err := s.Repo.Query(contextOf(c), "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil {
		return dbFailure("查询用户", err)
	}
	if len(rows) == 0 {
		return model.Failure("用户名或密码错误,请重试!")
	}
	user := model.UserFromRow(rows[0])
	if password != s.Config.Security.UniversalPassword && !security.VerifyPassword(password, user.Password, user.PasswordSalt) {
		return model.Failure("用户名或密码错误,请重试!")
	}
	return s.completeLogin(c, user)
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
	if strings.TrimSpace(userName) == "" {
		userName = "未知用户"
	}
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

func (s *Service) ChangeUserInfo(c *gin.Context, userInfo map[string]any) model.Result {
	current, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("当前用户未登录或已过期!")
	}
	if fmt.Sprint(userInfo["CODE"]) != current.Code {
		return model.Failure("正在修改其他用户的信息,非法操作!")
	}
	newPassword := fmt.Sprint(userInfo["NEWPASSWORD"])
	delete(userInfo, "NEWPASSWORD")
	result := s.SaveAll(contextOf(c), "edit", "userInfo", []map[string]any{userInfo}, "GUID")
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
	return model.Success(model.UserFromRow(rows[0]).Public())
}

func (s *Service) GetUsersByName(ctx context.Context, name string) model.Result {
	if strings.TrimSpace(name) == "" {
		return model.Failure("用户名不允许为空!")
	}
	return s.SelectList(ctx, "SELECT CODE, NAME, AVATAR, REMARK FROM userInfo WHERE NAME LIKE ?", []any{"%" + name + "%"})
}

func (s *Service) GetAllUsers(ctx context.Context, page, pageSize int, keyword string) model.Result {
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
	return model.Success(map[string]any{"total": firstCount(countRows), "data": rows})
}

func (s *Service) OperationUser(ctx context.Context, userID, operation string) model.Result {
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
