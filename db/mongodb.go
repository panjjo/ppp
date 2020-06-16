package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	log "github.com/sirupsen/logrus"
)

var (
	mongoClient *mongo.Client
	database    string
)
var (
	ErrObjNotArray    = fmt.Errorf("对象不是数组")
	ErrRecordNouFound = mongo.ErrNoDocuments
)

type M = bson.M

type Config struct {
	Addr   string `json:"addr" yaml:"addr" mapstructure:"addr"`
	DBName string `json:"db" yaml:"db" mapstructure:"db`
	Retry  *bool  `json:"retry" yaml:"retry"`
}

type Client struct {
	client *mongo.Client
	db     string
	ctx    context.Context
}

func InitDB(config Config) {
	log.Debugf("%+v", config)
	database = config.DBName
	ops := options.Client()
	if config.Retry != nil {
		ops.SetRetryWrites(*config.Retry)
	}
	ops.ApplyURI(config.Addr)

	client, err := mongo.NewClient(ops)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	mongoClient = client
	log.Infoln("Init Mogo")
}

func NewClient() *Client {
	return &Client{client: mongoClient, db: database, ctx: context.Background()}
}
func (c *Client) SetDB(db string) *Client {
	c.db = db
	return c
}
func (c *Client) SetSession(session mongo.SessionContext) *Client {
	c.ctx = session
	return c
}

func (c *Client) Insert(tb string, obj interface{}) error {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoInsert][%v][%s]\n", time.Since(now), tb)
	}()
	cl := c.client.Database(c.db).Collection(tb)
	_, err := cl.InsertOne(c.ctx, obj)
	return err
}

// Update 更新
func (c *Client) Update(tb string, query, update interface{}) error {
	_, _, err := c.UpdateResult(tb, query, update)
	return err
}

// UpdateResult 更新一条数据并返回 找到的数量，更新的数量
func (c *Client) UpdateResult(tb string, query, update interface{}) (int64, int64, error) {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoUpdateResult][%v][%s]\n", time.Since(now), tb)
	}()
	cl := c.client.Database(c.db).Collection(tb)
	r, err := cl.UpdateOne(c.ctx, query, update)
	if err != nil {
		return 0, 0, err
	}
	return r.MatchedCount, r.ModifiedCount, err
}

// Upsert 更新或插入，返回true表示更新
func (c *Client) Upsert(tb string, query, update interface{}) (bool, error) {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoUpsert][%v][%s]\n", time.Since(now), tb)
	}()
	cl := c.client.Database(c.db).Collection(tb)
	upsert := true
	r, err := cl.UpdateOne(context.Background(), query, update, &options.UpdateOptions{
		Upsert: &upsert,
	})
	if err != nil {
		return false, err
	}
	return r.MatchedCount == 0, err
}

// UpdateMany 更新多个记录 返回 修改数量,匹配数量,错误
func (c *Client) UpdateMany(tb string, query, update interface{}) (int64, int64, error) {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoUpdateMany][%v][%s]\n", time.Since(now), tb)
	}()
	cl := c.client.Database(c.db).Collection(tb)
	r, err := cl.UpdateMany(c.ctx, query, update)
	if err != nil {
		return 0, 0, err
	}
	return r.ModifiedCount, r.MatchedCount, nil
}

func (c *Client) Get(tb string, query, obj interface{}) error {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoGet][%v][%s]\n", time.Since(now), tb)
	}()
	cl := c.client.Database(c.db).Collection(tb)
	return cl.FindOne(c.ctx, query).Decode(obj)
}

func (c *Client) Count(tb string, query interface{}) (int64, error) {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoCount][%v][%s]\n", time.Since(now), tb)
	}()
	return c.client.Database(c.db).Collection(tb).CountDocuments(c.ctx, query)
}
func (c *Client) Find(tb string, query interface{}, skip, limit int64, sort string, total bool, obj interface{}) (int64, error) {
	now := time.Now()
	defer func() {
		log.Debugf("[MgoFind][%v][%s][%d:%d][%s]\n", time.Since(now), tb, skip, limit, sort)
	}()
	t := int64(0)
	rV := reflect.ValueOf(obj)
	if rV.Elem().Kind() != reflect.Slice {
		return t, ErrObjNotArray
	}

	cl := c.client.Database(c.db).Collection(tb)
	if total {
		tt, err := cl.CountDocuments(c.ctx, query)
		if err != nil {
			return t, err
		}
		t = tt
	}
	opt := options.Find()
	if skip > -1 {
		opt.Skip = &skip
	}
	if limit > -1 {
		opt.Limit = &limit
	}
	if sort != "" {
		sorts := strings.Split(sort, ",")
		sm := map[string]interface{}{}
		for _, v := range sorts {
			if v == "" {
				continue
			}
			if strings.HasPrefix(v, "-") {
				v = strings.Trim(v, "-")
				sm[v] = -1
			} else {
				sm[v] = 1
			}

		}
		if len(sm) > 0 {
			opt.Sort = sm
		}
	}

	cur, err := cl.Find(c.ctx, query, opt)
	if err != nil {
		return t, err
	}

	defer cur.Close(context.Background())
	slicev := rV.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for cur.Next(context.Background()) {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			err := cur.Decode(elemp.Interface())
			if err != nil {
				log.Fatal(err)
			}
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			err := cur.Decode(slicev.Index(i).Addr().Interface())
			if err != nil {
				log.Fatal(err)
			}
		}
		i++
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	rV.Elem().Set(slicev.Slice(0, i))
	return t, nil

}

func (c *Client) Del(tb string, query interface{}) (*mongo.DeleteResult, error) {
	log.Debugf("[MgoDel][%s]\n", tb)
	cl := c.client.Database(c.db).Collection(tb)
	return cl.DeleteOne(c.ctx, query)
}
func (c *Client) DelMany(tb string, query interface{}) (*mongo.DeleteResult, error) {
	log.Debugf("[MgoDel][%s][%s]\n", tb)
	cl := c.client.Database(c.db).Collection(tb)
	return cl.DeleteMany(c.ctx, query)
}
