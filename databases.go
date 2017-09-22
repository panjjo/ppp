package ppp

import (
	"fmt"

	"github.com/panjjo/ppp/pool"
	"gopkg.in/mgo.v2/bson"
)

var DBPool *pool.Pool

//新增用户
func saveUser(user User) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("user").Insert(user)
}

//获取用户信息
func getUser(userid, t string) User {
	session := DBPool.Get()
	defer session.Close()
	user := User{}
	session.DB().C("user").Find(bson.M{"userid": userid, "type": t}).One(&user)
	fmt.Println(user)
	return user
}

//更新用户信息
func updateUser(userid, t string, update bson.M) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("user").Update(bson.M{"userid": userid, "type": t}, update)
}
