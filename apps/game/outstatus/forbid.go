package outstatus

import (
	"kinger/gopuppy/apps/logic"
	"kinger/common/consts"
	"kinger/proto/pb"
	"strconv"
	"strings"
)

type iForbid interface {
	iClientOutStatus
	GetForbidID() int
}

type baseForbid struct {
	clientStatus
	forbidID int
}

func newForbid(statusID string) iOutStatus {
	if !strings.HasPrefix(statusID, consts.OtForbid) {
		return nil
	}
	info := strings.Split(statusID, "_")
	if len(info) < 2 {
		return nil
	}

	forbidID, _ := strconv.Atoi(info[1])
	switch forbidID {
	case consts.ForbidAccount:
		return &accountForbid{}
	case consts.ForbidChat:
		return &chatForbid{}
	case consts.ForbidMonitor:
		return &monitorForbid{}
	default:
		return nil
	}
}

func (b *baseForbid) GetForbidID() int {
	if b.forbidID <= 0 {
		info := strings.Split(b.GetID(), "_")
		if len(info) < 2 {
			return 0
		}
		b.forbidID, _ = strconv.Atoi(info[1])
	}
	return b.forbidID
}

type accountForbid struct {
	baseForbid
}


type chatForbid struct {
	baseForbid
}

func (cf *chatForbid) onDel()() {
	logic.BroadcastBackend(pb.MessageID_L2CA_FORBID_CHAT, &pb.ForbidChatArg{
		Uid:      uint64(cf.player.GetUid()),
		IsForbid: false,
	})
}

type monitorForbid struct {
	baseForbid
}


