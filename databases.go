package ppp

import (
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
func getUser(userid string, t string) User {
	return getUserQuery(bson.M{"userid": userid, "type": t})
}
func getUserQuery(q bson.M) User {
	session := DBPool.Get()
	defer session.Close()
	user := User{}
	session.DB().C("user").Find(q).One(&user)
	return user
}

//更新用户信息
func updateUser(userid string, t string, update bson.M) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("user").Update(bson.M{"userid": userid, "type": t}, update)
}

//更新用户信息
func updateUserMulti(query, update bson.M) error {
	session := DBPool.Get()
	defer session.Close()
	_, err := session.DB().C("user").UpdateAll(query, update)
	return err
}

//获取授权
func getToken(mchid, t string) authBase {
	session := DBPool.Get()
	defer session.Close()
	auth := authBase{}
	session.DB().C("authinfo").Find(bson.M{"mchid": mchid, "type": t}).One(&auth)
	return auth
}

//刷新授权
func updateToken(mchid, t string, update bson.M) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("authinfo").Update(bson.M{"mchid": mchid, "type": t}, update)
}

//保存授权
func saveToken(auth authBase) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("authinfo").Insert(auth)
}

//删除授权
func deleteToken(mchid, t string) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("authinfo").Remove(bson.M{"mchid": mchid, "type": t})
}

//获取交易
func getTrade(q bson.M) Trade {
	session := DBPool.Get()
	defer session.Close()
	trade := Trade{}
	session.DB().C("trade").Find(q).One(&trade)
	return trade
}

//新增交易
func saveTrade(trade Trade) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("trade").Insert(trade)
}
func updateTrade(query, update bson.M) error {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("trade").Update(query, update)
}

func listTrade(request *ListRequest) ([]Trade, error) {
	session := DBPool.Get()
	defer session.Close()
	trades := []Trade{}
	col := session.DB().C("trade").Find(request.Query)
	if request.Limit != 0 {
		col = col.Limit(request.Limit)
	}
	if request.Skip > 0 {
		col = col.Skip(request.Skip)
	}
	if request.Sort != "" {
		col = col.Sort(request.Sort)
	}
	err := col.All(&trades)
	return trades, err
}

func countTrade(request *ListRequest) (int, error) {
	session := DBPool.Get()
	defer session.Close()
	return session.DB().C("trade").Find(request.Query).Count()
}
