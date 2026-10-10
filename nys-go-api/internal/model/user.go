package model

// User keeps the original uppercase field names so existing Vue pages remain compatible.
type User struct {
	GUID         string `json:"guid,omitempty"`
	Code         string `json:"code,omitempty"`
	Name         string `json:"name,omitempty"`
	Password     string `json:"password,omitempty"`
	PasswordSalt string `json:"passwordsalt,omitempty"`
	Role         string `json:"role,omitempty"`
	Remark       string `json:"remark,omitempty"`
	LoginAddress string `json:"loginaddress,omitempty"`
	LoginIP      string `json:"loginip,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	IsBan        string `json:"isban,omitempty"`
	UserNum      int64  `json:"usernum"`
}

func UserFromRow(row map[string]any) User {
	return User{
		GUID:         StringValue(row, "GUID"),
		Code:         StringValue(row, "CODE"),
		Name:         StringValue(row, "NAME"),
		Password:     StringValue(row, "PASSWORD"),
		PasswordSalt: StringValue(row, "PASSWORDSALT"),
		Role:         StringValue(row, "ROLE"),
		Remark:       StringValue(row, "REMARK"),
		LoginAddress: StringValue(row, "LOGINADDRESS"),
		LoginIP:      StringValue(row, "LOGINIP"),
		Phone:        StringValue(row, "PHONE"),
		Email:        StringValue(row, "EMAIL"),
		Avatar:       StringValue(row, "AVATAR"),
		IsBan:        StringValue(row, "ISBAN"),
		UserNum:      Int64Value(row, "USERNUM"),
	}
}

func (u User) Public() User {
	u.Password = ""
	u.PasswordSalt = ""
	u.LoginIP = ""
	return u
}

// PublicProfile 用于查看他人主页等无需登录的接口：在 Public 的基础上再去掉手机号。
// 邮箱和登录地区是个人主页上公开展示的信息，予以保留。
func (u User) PublicProfile() User {
	u = u.Public()
	u.Phone = ""
	return u
}
