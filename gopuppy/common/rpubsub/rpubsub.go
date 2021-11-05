package rpubsub

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
)

var redisClient *redis.Client

func Initialize(addr string) {
	redisClient = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    addr,
	})
}

type Handler func(map[string]interface{})

func SubscribeWithOption(channel string, handler Handler, isSync bool) error {
	var rps *redis.PubSub
	var err error
	if !isSync {
		evq.Await(func() {
			rps, err = redisClient.Subscribe(channel)
		})
	} else {
		rps, err = redisClient.Subscribe(channel)
	}
	if err != nil {
		return err
	}

	go func() {
		for {
			utils.CatchPanic(func() {
				msg, err := rps.ReceiveMessage()
				if err != nil {
					glog.Errorf("rpubSub.ReceiveMessage err =%s", err)
					return
				}

				if msg.Channel == channel {
					msgInfo := map[string]interface{}{}
					err := json.Unmarshal([]byte(msg.Payload), &msgInfo)
					if err != nil {
						glog.Infof("rpubSub.ReceiveMessage Unmarshal err =%s", err)
					}

					evq.CallLater(func() {
						handler(msgInfo)
					})
				} else {
					glog.Infof("rpubSub.ReceiveMessage msg =%s", msg)
				}
			})
		}
	}()

	return nil
}

func Subscribe(channel string, handler Handler) error {
	return SubscribeWithOption(channel, handler, false)
}

func Publish(channel string, message map[string]interface{}) error {
	var err error
	evq.Await(func() {
		payload, err2 := json.Marshal(message)
		if err2 != nil {
			err = err2
			return
		}

		redisClient.Publish(channel, string(payload))
	})
	return err
}
