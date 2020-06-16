package ppp

import "github.com/panjjo/ppp/db"

const (
	userTable = "users"

	//UserWaitVerify 用户等待审核或等待授权签约
	UserWaitVerify Status = 0
	//UserFreeze 用户冻结
	UserFreeze Status = -1
	//UserSucc 用户正常
	UserSucc Status = 1
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
	user := &User{}
	DBClient.Get(userTable, db.M{"userid": userid, "from": t}, user)
	return user
}

// 更新用户
func updateUser(query, update interface{}) error {
	return DBClient.Update(userTable, query, db.M{"$set": update})
}

//
func updateUserMulti(query, update interface{}) error {
	_, _, err := DBClient.UpdateMany(userTable, query, db.M{"$set": update})
	return err
}

// 保存用户
func saveUser(user *User) error {
	return DBClient.Insert(userTable, user)
}
