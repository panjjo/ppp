package db

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// Config 数据库配置文件
type Config struct {
	Addr      string `yaml:"addr"`
	DB        string `yaml:"db"`
	MaxActive int    `yaml:"maxActive"`
	Type      string `yaml:"type"`
}

// Conn 数据库连接接口
// 默认使用的是mongodb，可自行实现对应接口更换数据库
type Conn interface {
	// Ping 检查连接是否可用
	Ping() error
	// Close 关闭连接
	Close() error

	p(*Pool)
	time() time.Time
	setTime(time.Time)

	// FindOne 查询一个
	FindOne(tb string, query, res interface{}) interface{}
	// Update 更新默认全部
	Update(tb string, query, update interface{}) error
	// Save 保存
	Save(tb string, data interface{}) error
	// UpSert 存在更新，不存在新增
	UpSert(tb string, query, update interface{}) (interface{}, error)
	// updateAll 批量更新
	UpAll(tb string, query, update interface{}) (interface{}, error)
}

// Pool maintains a pool of connections. The application calls the Get method
// to get a connection from the pool and the connection's Close method to
// return the connection's resources to the pool.
type Pool struct {
	Dial func() (Conn, error)

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c Conn) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool

	// mu protects fields defined below.
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
	active int

	// Stack of idleConn with most recently used at the front.
	idle list.List
}

// Get gets a connection. The application must close the returned connection.
// This method always returns a valid connection so that applications can defer
// error handling to the first use of the connection. If there is an error
// getting an underlying connection, then the connection Err, Do, Send, Flush
// and Receive methods return that error.
func (p *Pool) Get() Conn {
	c, err := p.get()
	if err != nil {
		return nil
	}
	c.p(p)
	return c
}

// ActiveCount returns the number of active connections in the pool.
func (p *Pool) ActiveCount() int {
	p.mu.Lock()
	active := p.active
	p.mu.Unlock()
	return active
}

// Close releases the resources used by the pool.
func (p *Pool) Close() error {
	p.mu.Lock()
	idle := p.idle
	p.idle.Init()
	p.closed = true
	p.active -= idle.Len()
	if p.cond != nil {
		p.cond.Broadcast()
	}
	p.mu.Unlock()
	for e := idle.Front(); e != nil; e = e.Next() {
		e.Value.(Conn).Close()
	}
	return nil
}

// release decrements the active count and signals waiters. The caller must
// hold p.mu during the call.
func (p *Pool) release() {
	p.active--
	if p.cond != nil {
		p.cond.Signal()
	}
}

// get prunes stale connections and returns a connection from the idle list or
// creates a new connection.
func (p *Pool) get() (Conn, error) {
	p.mu.Lock()

	// Prune stale connections.
	if timeout := p.IdleTimeout; timeout > 0 {
		for i, n := 0, p.idle.Len(); i < n; i++ {
			e := p.idle.Back()
			if e == nil {
				break
			}
			ic := e.Value.(Conn)
			if ic.time().Add(timeout).After(time.Now()) {
				break
			}
			p.idle.Remove(e)
			p.release()
			p.mu.Unlock()
			ic.Close()
			p.mu.Lock()
		}
	}

	for {

		// Get idle connection.
		for i, n := 0, p.idle.Len(); i < n; i++ {
			e := p.idle.Front()
			if e == nil {
				break
			}
			ic := e.Value.(Conn)
			p.idle.Remove(e)
			test := p.TestOnBorrow
			p.mu.Unlock()
			if test == nil || test(ic) == nil {
				return ic, nil
			}
			ic.Close()
			p.mu.Lock()
			p.release()
		}

		// Check for pool closed before dialing a new connection.
		if p.closed {
			p.mu.Unlock()
			return nil, errors.New("flysnow: get on closed pool")
		}

		// Dial new connection if under limit.
		if p.MaxActive == 0 || p.active < p.MaxActive {
			p.active++
			p.mu.Unlock()
			c, err := p.Dial()
			if err != nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				c = nil
			}
			return c, err
		}

		if !p.Wait {
			p.mu.Unlock()
			return nil, errors.New("flysnow: connection pool exhausted")
		}

		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}
		p.cond.Wait()
	}
}

func (p *Pool) put(c Conn) error {
	p.mu.Lock()
	if !p.closed {
		c.setTime(time.Now())
		p.idle.PushFront(c)
		if p.idle.Len() > p.MaxIdle {
			c = p.idle.Remove(p.idle.Back()).(Conn)
		} else {
			c = nil
		}
	}

	if c == nil {
		if p.cond != nil {
			p.cond.Signal()
		}
		p.mu.Unlock()
		return nil
	}

	p.release()
	p.mu.Unlock()
	c.Close()
	return nil
}
