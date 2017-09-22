package pool

import (
	"time"

	"gopkg.in/mgo.v2"
)

type Config struct {
	Addr      string
	Port      int
	DB        string
	MaxActive int
}

type Conn struct {
	conn *mgo.Session
	db   *mgo.Database

	p *Pool
	t time.Time
}

func (c *Conn) Ping() error {
	return c.conn.Ping()
}
func (f *Conn) Close() error {
	f.p.put(f)
	return nil
}
func (f *Conn) DB() *mgo.Database {
	return f.db
}

var mongoSession *mgo.Session

func NewConnection(config *Config) (*Conn, error) {

	if mongoSession == nil {
		session, err := mgo.Dial(config.Addr)
		if err != nil {
			return nil, err
		}
		mongoSession = session
	}
	tmpsess := mongoSession.Clone()
	return &Conn{
		conn: tmpsess,
		db:   tmpsess.DB(config.DB),
		t:    time.Now(),
	}, nil
}
func GetPool(config *Config) *Pool {
	return &Pool{
		Dial: func() (*Conn, error) { return NewConnection(config) },
		TestOnBorrow: func(c *Conn) error {
			return c.Ping()
		},
		MaxIdle:     config.MaxActive,
		MaxActive:   config.MaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
	}
}
