package ppp2

const (
	userTable = "users"
)

// User 用户信息
// 关联着用户和授权，存在多个用户使用同一个授权的情况
type User struct {
	UserID string // 外部商户的ID
	ID     string
	MchID  string // 第三方账号Auth.MchID
	From   string
	Status Status
}

// 查询用户
func getUser(userid, t string) *User {
	session := DBPool.Get()
	defer session.Close()
	user := &User{}
	res := session.FindOne(userTable, map[string]interface{}{"userid": userid, "from": t}, user)
	if res != nil {
		user = res.(*User)
	}
	return user
}

// 更新用户
func updateUser(query, update interface{}) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Update(userTable, query, update)
}

// 保存用户
func saveUser(user *User) error {
	session := DBPool.Get()
	defer session.Close()
	return session.Save(userTable, user)
}
