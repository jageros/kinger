package mq

import (
	"github.com/pkg/errors"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/proto/pb"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/network"
	"kinger/gopuppy/common/evq"
)

var (
	rmqMessgeGreaters = map[int32]func() IRmqMessge{}
	consumers = map[string]IConsumer {}
)

func RegisterRmqMessage(type_ protoc.IMessageID, greater func() IRmqMessge) {
	rmqMessgeGreaters[type_.ID()] = greater
}

type IRmqMessge interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type IConsumer interface {
	Consume(type_ int32, msg IRmqMessge)
}

func decodeRmqMessage(msg *pb.RmqMessage) (IRmqMessge, error) {
	if greater, ok := rmqMessgeGreaters[msg.Type]; ok {
		rmqMsg := greater()
		if rmqMsg == nil {
			return nil, errors.Errorf("greate nil rmqMsg %s", msg.Type)
		}

		err := rmqMsg.Unmarshal(msg.Payload)
		if err != nil {
			return nil, err
		}
		return rmqMsg, nil
	} else {
		return nil, errors.Errorf("%s no greater", msg.Type)
	}
}


func Publish(queue string, type_ protoc.IMessageID, region uint32, msg IRmqMessge) error {
	payload, err := msg.Marshal()
	if err != nil {
		glog.Errorf("Publish encodeRmqMessage err %d", err)
		return err
	}

	api.SelectCenterByString(queue, region).Push(pb.MessageID_L2C_MQ_PUBLISH, &pb.RmqMessage{
		Queue: queue,
		Type: type_.ID(),
		Payload: payload,
	})
	return nil
}

func AddConsumer(queue string, region uint32, consumer IConsumer) error {
	if _, ok := consumers[queue]; ok {
		glog.Warnf("%s AddConsumer repeat", queue)
		consumers[queue] = consumer
		return nil
	}

	consumers[queue] = consumer
	api.SelectCenterByString(queue, region).Push(pb.MessageID_L2C_MQ_ADD_CONSUMER, &pb.MqConsumerArg{
		Queue: queue,
	})
	return nil
}

func RemoveConsumer(queue string, region uint32) {
	//glog.Infof("onPlayerKickOut RemoveConsumer 11111111111")
	if _, ok := consumers[queue]; !ok {
		return
	}
	//glog.Infof("onPlayerKickOut RemoveConsumer 22222222222")
	api.SelectCenterByString(queue, region).Push(pb.MessageID_L2C_MQ_REMOVE_CONSUMER, &pb.MqConsumerArg{
		Queue: queue,
	})
	delete(consumers, queue)
}

func rpc_C2L_MqConsume(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RmqMessage)
	consumer, ok := consumers[arg2.Queue]
	if ok {
		evq.CallLater(func() {
			msg, err := decodeRmqMessage(arg2)
			if err != nil {
				glog.Errorf("decodeRmqMessage err, type=%d", arg2.Type)
				return
			}
			consumer.Consume(arg2.Type, msg)
		})
		return nil, nil
	} else {
		glog.Errorf("rpc_C2L_MqConsume no consumer, queue=%s", arg2.Queue)
		return nil, network.InternalErr
	}
}

func InitClient() {
	api.RegisterCenterRpcHandler(pb.MessageID_C2L_MQ_CONSUME, rpc_C2L_MqConsume)
}
