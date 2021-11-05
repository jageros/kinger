package redis

import (
	"github.com/gomodule/redigo/redis"
	"kinger/gopuppy/common/async"
)

const (
	_REDIS_ASYNC_JOB_GROUP = "_redis"
)

type DB struct {
	Conn redis.Conn
}

func Dial(network, address string, options []redis.DialOption) (*DB, error) {
	//async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
	conn, err := redis.Dial(network, address, options...)
	if err == nil {
		return &DB{conn}, nil
	} else {
		return nil, err
	}
	//}, ac)
}

func DialURL(rawurl string, options []redis.DialOption) (*DB, error) {
	//async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
	conn, err := redis.DialURL(rawurl, options...)
	if err == nil {
		return &DB{conn}, nil
	} else {
		return nil, err
	}
	//}, ac)
}

func (db *DB) Do(commandName string, args []interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		res, err = db.Conn.Do(commandName, args...)
		return
	}, ac)
}
