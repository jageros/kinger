package rediscluster

import (
	redis "github.com/chasex/redis-go-cluster"
	"github.com/pkg/errors"
	"kinger/gopuppy/common/async"
	"time"
)

var DB *RedisClusterEngine

const (
	_REDIS_CLUSTER_ASYNC_JOB_GROUP = "_redis_cluster"
)

type RedisClusterEngine struct {
	c *redis.Cluster
}

func OpenRedisCluster(startNodes []string) (*RedisClusterEngine, error) {
	c, err := redis.NewCluster(&redis.Options{
		StartNodes:   startNodes,
		ConnTimeout:  10 * time.Second, // Connection timeout
		ReadTimeout:  60 * time.Second, // Read timeout
		WriteTimeout: 60 * time.Second, // Write timeout
		KeepAlive:    1,                // Maximum keep alive connecion in each node
		AliveTime:    10 * time.Minute, // Keep alive timeout
	})

	if err != nil {
		return nil, errors.Wrap(err, "connect redis cluster failed")
	}

	DB = &RedisClusterEngine{
		c: c,
	}

	return DB, nil
}

func (es *RedisClusterEngine) Do(commandName string, args []interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_CLUSTER_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		res, err = es.c.Do(commandName, args...)
		return
	}, ac)
}

func (es *RedisClusterEngine) Send(commandName string, args []interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_CLUSTER_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		res, err = es.c.Do(commandName, args...)
		return
	}, ac)
}

func (es *RedisClusterEngine) Close() {
	es.c.Close()
}
