package common

import (
	"fmt"
	"github.com/edwingeng/slog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/glog"

	mongoWuid "github.com/edwingeng/wuid/mongo/wuid"
	"kinger/gopuppy/common/timer"
	"time"
)

var (
	uuidGenerators     = make(map[string]*mongoWuid.WUID)
	uuid32MongoSession *mgo.Session
)

type wuidLog struct{}

func (l *wuidLog) NewLoggerWith(keyVals ...interface{}) slog.Logger {
	return &wuidLog{}
}

func (l *wuidLog) Debug(args ...interface{}) {
	glog.Debugf("%v", args)
}
func (l *wuidLog) Info(args ...interface{}) {
	glog.Infof("%v", args)
}
func (l *wuidLog) Warn(args ...interface{}) {
	glog.Warnf("%v", args)
}
func (l *wuidLog) Error(args ...interface{}) {
	glog.Error(args...)
}

func (l *wuidLog) Debugf(template string, args ...interface{}) {
	glog.Debugf(template, args...)
}
func (l *wuidLog) Infof(template string, args ...interface{}) {
	glog.Infof(template, args...)
}
func (l *wuidLog) Warnf(template string, args ...interface{}) {
	glog.Warnf(template, args...)
}
func (l *wuidLog) Errorf(template string, args ...interface{}) {
	glog.Errorf(template, args...)
}

func (l *wuidLog) Debugw(msg string, keyVals ...interface{}) {
	glog.Debugf("msg=%s args=%v", msg, keyVals)
}
func (l *wuidLog) Infow(msg string, keyVals ...interface{}) {
	glog.Infof("msg=%s args=%v", msg, keyVals)
}
func (l *wuidLog) Warnw(msg string, keyVals ...interface{}) {
	glog.Warnf("msg=%s args=%v", msg, keyVals)
}
func (l *wuidLog) Errorw(msg string, keyVals ...interface{}) {
	glog.Errorf("msg=%s args=%v", msg, keyVals)
}

func (l *wuidLog) FlushLogger() error {
	return glog.Flush()
}

// =============

func InitUUidGenerator(tag string) *mongoWuid.WUID {
	generator, ok := uuidGenerators[tag]
	if !ok {
		generator = mongoWuid.NewWUID(tag, &wuidLog{})
		cfg := config.GetGlobalDbConfig()
		for {
			if cfg.Type == "mongodb" {
				//cfg.Addr, cfg.User, cfg.Password, cfg.DB, "uuid", tag
				if err := generator.LoadH28FromMongo(func() (client *mongo.Client, autoDisconnect bool, err error) {
					cli, err := mongo.NewClient(&options.ClientOptions{
						Auth: &options.Credential{
							Username: cfg.User,
							Password: cfg.Password,
						},
						Hosts: []string{cfg.Addr},
					})
					if err != nil {
						return nil, true, err
					}
					return cli, true, nil
				}, cfg.DB, "uuid", tag); err != nil {
					glog.Errorf("LoadH24FromMongo %s %v %s", tag, cfg, err)
					continue
				} else {
					break
				}
			} else {
				panic(fmt.Sprintf("InitUUidGenerator unknow db type %s", cfg.Type))
			}
		}
		uuidGenerators[tag] = generator
	}
	return generator
}

func Init32UUidGenerator() {
	cfg := config.GetGlobalDbConfig()
	var url = "mongodb://" + cfg.Addr + "/"
	var err error
	mongoSes, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}

	uuid32MongoSession = mongoSes
	db := mongoSes.DB(cfg.DB)
	if len(cfg.User) > 0 {
		if err = db.Login(cfg.User, cfg.Password); err != nil {
			panic(err)
		}
	}

	timer.AddTicker(10*time.Second, func() {
		mongoSes.Ping()
	})
}

type UUid uint64

func GenUUid(tag string) UUid {
	return UUid(InitUUidGenerator(tag).Next())
}

func (u UUid) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

func ParseUUidFromString(val string) UUid {
	intVal, _ := strconv.ParseUint(val, 10, 64)
	return UUid(intVal)
}

func Gen32UUid(tag string) (uint32, error) {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"n": 1}},
		Upsert:    true,
		ReturnNew: true,
	}

	ses := uuid32MongoSession.Copy()
	db := ses.DB(config.GetGlobalDbConfig().DB)
	c := db.C("uuid")
	m := make(map[string]interface{})
	_, err := c.FindId(tag).Apply(change, &m)
	ses.Close()

	if err != nil {
		return 0, err
	}

	return uint32(m["n"].(int)), nil
}
