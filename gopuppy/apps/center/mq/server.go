package mq

import (
	"gopkg.in/redis.v3"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/gopuppy/network"
	"kinger/gopuppy/proto/pb"
	"time"
)

var (
	consumer2Queues = map[int64]common.StringSet{}
	queue2Consumer  = map[string]*network.Session{}
	redisClient     *redis.Client
)

func rpc_L2C_MqAddConsumer(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.MqConsumerArg)
	if arg2.Queue == "" {
		return nil, network.InternalErr
	}

	sesID := ses.GetSesID()
	queues, ok := consumer2Queues[sesID]
	if !ok {
		queues = common.StringSet{}
		consumer2Queues[sesID] = queues
	}
	queues.Add(arg2.Queue)

	if oldSes, ok := queue2Consumer[arg2.Queue]; ok && oldSes != ses {
		glog.Warnf("rpc_L2C_MqAddConsumer queue %s has Consumer %d", arg2.Queue, oldSes.GetSesID())
	}
	queue2Consumer[arg2.Queue] = ses

	var msgs []*pb.RmqMessage
	evq.Await(func() {
		for {
			data, err := redisClient.LPop(arg2.Queue).Bytes()
			if err != nil || data == nil || len(data) == 0 {
				break
			}

			msg := &pb.RmqMessage{}
			err = msg.Unmarshal(data)
			if err == nil {
				msgs = append(msgs, msg)
			}
		}
	})

	for _, msg := range msgs {
		onPublish(msg)
	}

	return nil, nil
}

func rpc_L2C_MqRemoveConsumer(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.MqConsumerArg)
	if arg2.Queue == "" {
		return nil, network.InternalErr
	}

	sesID := ses.GetSesID()
	queues, ok := consumer2Queues[sesID]
	if ok {
		queues.Remove(arg2.Queue)
	}

	if oldSes, ok := queue2Consumer[arg2.Queue]; ok && oldSes == ses {
		delete(queue2Consumer, arg2.Queue)
	}
	return nil, nil
}

func rpc_L2C_MqPublish(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RmqMessage)
	if arg2.Queue == "" {
		return nil, network.InternalErr
	}

	onPublish(arg2)
	return nil, nil
}

func onPublish(msg *pb.RmqMessage) {
	if consumer, ok := queue2Consumer[msg.Queue]; ok {
		_, err := consumer.Call(pb.MessageID_C2L_MQ_CONSUME, msg)
		if err != nil {
			timer.AfterFunc(200*time.Millisecond, func() {
				onPublish(msg)
			})
		}
	} else {
		evq.Await(func() {
			data, _ := msg.Marshal()
			redisClient.RPush(msg.Queue, string(data))
		})
	}
}

func onSessionClose(ses *network.Session) {
	sesID := ses.GetSesID()
	queues, ok := consumer2Queues[sesID]
	if !ok {
		return
	}
	delete(consumer2Queues, sesID)

	queues.ForEach(func(q string) bool {
		if oldSes, ok := queue2Consumer[q]; ok && oldSes == ses {
			delete(queue2Consumer, q)
		}
		return true
	})
}

func InitServer(peer *network.Peer) {
	evq.HandleEvent(consts.SESSION_ON_CLOSE_EVENT, func(event evq.IEvent) {
		ses := event.(*evq.CommonEvent).GetData()[0].(*network.Session)
		if ses.FromPeer() == peer {
			onSessionClose(ses)
		}
	})

	redisClient = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.GetRegionConfig().Redis.Addr,
	})

	go func() {
		for {
			time.Sleep(20 * time.Second)
			utils.CatchPanic(func() {
				redisClient.Ping()
			})
		}
	}()

	peer.RegisterRpcHandler(pb.MessageID_L2C_MQ_ADD_CONSUMER, rpc_L2C_MqAddConsumer)
	peer.RegisterRpcHandler(pb.MessageID_L2C_MQ_REMOVE_CONSUMER, rpc_L2C_MqRemoveConsumer)
	peer.RegisterRpcHandler(pb.MessageID_L2C_MQ_PUBLISH, rpc_L2C_MqPublish)
}
