package db

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MgoConn mongodb的连接实例
type MgoConn struct {
	conn *mgo.Session
	_db  *mgo.Database

	pool *Pool
	t    time.Time
}

// Ping 实现接口
func (c *MgoConn) Ping() error {
	return c.conn.Ping()
}

// Close 实现接口
func (c *MgoConn) Close() error {
	c.pool.put(c)
	return nil
}

// p 实现接口
func (c *MgoConn) p(pool *Pool) {
	c.pool = pool
}

// st 实现接口
func (c *MgoConn) setTime(t time.Time) {
	c.t = t
}

// st 实现接口
func (c *MgoConn) time() time.Time {
	return c.t
}

// FindOne 实现查询接口
func (c MgoConn) FindOne(tb string, query interface{}, res interface{}) interface{} {
	c._db.C(tb).Find(query).One(res)
	return res
}

// Update 实现查询接口
func (c MgoConn) Update(tb string, query, update interface{}) (err error) {
	return c._db.C(tb).Update(query, update)
}

// UpSert 实现有就更新，没有新增
func (c MgoConn) UpSert(tb string, query, update interface{}) (change interface{}, err error) {
	return c._db.C(tb).Upsert(query, update)
}

// UpAll 实现有就更新，没有新增
func (c MgoConn) UpAll(tb string, query, update interface{}) (change interface{}, err error) {
	return c._db.C(tb).UpdateAll(query, bson.M{"$set": update})
}

// Save 实现保存接口
func (c MgoConn) Save(tb string, data interface{}) error {
	return c._db.C(tb).Insert(data)
}

var mongoSession *mgo.Session

//NewMgoConnection 生成mongo 连接
func NewMgoConnection(config *Config) (Conn, error) {

	if mongoSession == nil {
		session, err := mgo.Dial(config.Addr)
		if err != nil {
			return nil, err
		}
		mongoSession = session
	}
	tmpsess := mongoSession.Clone()
	return &MgoConn{
		conn: tmpsess,
		_db:  tmpsess.DB(config.DB),
		t:    time.Now(),
	}, nil
}

// GetPool 生成连接池
func GetPool(config *Config) *Pool {
	return &Pool{
		// 更换数据库类型请替换 NewMgoConnection
		Dial: func() (Conn, error) { return NewMgoConnection(config) },
		TestOnBorrow: func(c Conn) error {
			return c.Ping()
		},
		MaxIdle:     config.MaxActive,
		MaxActive:   config.MaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
	}
}
