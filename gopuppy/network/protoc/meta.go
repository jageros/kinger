package protoc

var metaData = make(map[int32]IMeta)

type IMeta interface {
	GetMessageID() IMessageID
	EncodeArg(interface{}) ([]byte, error)
	DecodeArg([]byte) (interface{}, error)
	EncodeReply(interface{}) ([]byte, error)
	DecodeReply([]byte) (interface{}, error)
}

func RegisterMeta(meta IMeta) {
	metaData[meta.GetMessageID().ID()] = meta
}

func GetMeta(msgId int32) IMeta {
	if m, ok := metaData[msgId]; ok {
		return m
	} else {
		return nil
	}
}
