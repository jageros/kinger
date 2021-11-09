package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/snet"
	_ "kinger/meta"
	"kinger/proto/pb"
	"math/rand"
	"sync"
	"time"
)

var (
	gates [][]interface{} = [][]interface{}{
		[]interface{}{"134.175.13.251", 9104},
		//[]interface{}{"134.175.13.251", 9103},
	}

	gateIndex   = 0
	peer        *network.Peer
	closeSignal chan struct{}
	wait        sync.WaitGroup
	robotAmount        = 1000
	robotMinID  uint32 = 1
	id2Robot    map[uint32]*robot
)

type rpcCall struct {
	msgID pb.MessageID
	arg   interface{}
}

type robot struct {
	id        uint32
	uid       uint64
	gateIp    string
	gatePort  int
	ses       *network.Session
	rpcChan   chan *rpcCall
	camp      int
	battleObj *battle
}

func newRobot(id uint32, gateIp string, gatePort int) *robot {
	return &robot{
		id:       id,
		gateIp:   gateIp,
		gatePort: gatePort,
		rpcChan:  make(chan *rpcCall, 10),
	}
}

func (r *robot) login() {
	for {
		r.ses, _ = peer.DialWs(r.gateIp, r.gatePort, nil, &snet.Config{
			EnableCrypt:        false,
			HandshakeTimeout:   10 * time.Second,
			RewriterBufferSize: 100 * 1024,
			ReconnWaitTimeout:  5 * time.Minute,
		})

		c := r.ses.CallAsync(pb.MessageID_C2S_SDK_ACCOUNT_LOGIN, &pb.AccountLoginArg{
			Channel:   "robot",
			ChannelID: fmt.Sprintf("robot%d", r.id),
			Account:   fmt.Sprintf("robot%d", r.id),
		})
		result := <-c
		if result.Err != nil {
			<-r.ses.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		c = r.ses.CallAsync(pb.MessageID_C2S_LOGIN, &pb.LoginArg{
			Channel:     "robot",
			ArchiveID:   1,
			ChannelID:   fmt.Sprintf("robot%d", r.id),
			AccountType: pb.AccountTypeEnum_Ios,
		})
		result = <-c
		if result.Err != nil {
			<-r.ses.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		reply := result.Reply.(*pb.LoginReply)
		r.uid = reply.Uid
		camp := int(reply.GuideCamp)
		if reply.GuideCamp <= 0 {
			camps := []int{consts.Wei, consts.Shu, consts.Wu}
			camp = camps[rand.Intn(len(camps))]
			c = r.ses.CallAsync(pb.MessageID_C2S_SET_CAMP_ID, &pb.SetCampIDArg{
				CampID: int32(camp),
			})
			result = <-c
			if result.Err != nil {
				<-r.ses.Close()
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		r.ses.SetProp("id", r.id)
		r.camp = camp
		break
	}
}

func (r *robot) beginMatch() {
	t1 := time.Now()

	c := r.ses.CallAsync(pb.MessageID_C2S_BEGIN_MATCH, &pb.MatchArg{
		Camp: int32(r.camp),
	})
	result := <-c

	t2 := time.Now()
	d := t2.Sub(t1)
	glog.Infof("beginMatch id=%d, time=%s, err=%s", r.id, d, result.Err)

	if result.Err != nil {
		time.Sleep(100 * time.Millisecond)
		r.beginMatch()
	}
}

func (r *robot) onMatchDone(arg interface{}) {
	t1 := time.Now()

	arg2 := arg.(*pb.MatchInfo)
	c := r.ses.CallAsync(pb.MessageID_C2S_MATCH_READY_DONE, &pb.MatchDoneArg{
		RoomId: arg2.RoomId,
	})
	result := <-c

	t2 := time.Now()
	d := t2.Sub(t1)
	glog.Infof("onMatchDone id=%d, time=%s, err=%s", r.id, d, result.Err)

	r.beginMatch()
}

func (r *robot) beginLevelBattle() {
	t1 := time.Now()

	c := r.ses.CallAsync(pb.MessageID_C2S_BEGIN_LEVEL_BATTLE, &pb.BeginLevelBattle{
		LevelId: 36,
	})
	result := <-c

	t2 := time.Now()
	d := t2.Sub(t1)
	glog.Infof("beginLevelBattle id=%d, time=%s, err=%s", r.id, d, result.Err)

	time.Sleep(100 * time.Millisecond)
	if result.Err != nil {
		r.beginLevelBattle()
		return
	}

	r.battleObj = newBattle(r, result.Reply.(*pb.LevelBattle))

	t1 = time.Now()
	c = r.ses.CallAsync(pb.MessageID_C2S_LEVEL_READY_DONE, &pb.LevelChooseCard{})
	result = <-c
	t2 = time.Now()
	d = t2.Sub(t1)
	glog.Infof("C2S_LEVEL_READY_DONE id=%d, time=%s, err=%s", r.id, d, result.Err)

	if result.Err != nil {
		r.battleObj = nil
		r.beginLevelBattle()
		return
	}
}

func (r *robot) onBattleBoutBegin(arg interface{}) {
	if r.battleObj == nil {
		glog.Errorf("onBattleBoutBegin not battle")
		r.beginLevelBattle()
		return
	}

	arg2 := arg.(*pb.FightBoutBegin)
	//glog.Infof("onBattleBoutBegin %d", arg2.BoutUid)
	if arg2.BoutUid != r.uid {
		return
	}

	time.Sleep(3 * time.Second)
	r.battleObj.playCard()
}

func (r *robot) onBattleBoutResult(arg interface{}) {
	if r.battleObj == nil {
		glog.Infof("onBattleBoutResult not battle")
		r.beginLevelBattle()
		return
	}

	//glog.Infof("onBattleBoutResult")
	r.battleObj.onPlayCard(arg.(*pb.FightBoutResult))
}

func (r *robot) onBattleEnd(arg interface{}) {
	glog.Infof("onBattleEnd %d", r.id)
	r.battleObj = nil
	time.Sleep(time.Second)
	r.beginLevelBattle()
}

func (r *robot) fuckIt() {
	go func() {
		r.login()
		//r.beginMatch()
		r.beginLevelBattle()

	L1:
		for {

			select {
			case call := <-r.rpcChan:
				switch call.msgID {
				case pb.MessageID_S2C_MATCH_TIMEOUT:
					r.beginMatch()
				case pb.MessageID_S2C_UPDATE_MATCH_INFO:
					r.onMatchDone(call.arg)
				case pb.MessageID_S2C_FIGHT_BOUT_BEGIN:
					r.onBattleBoutBegin(call.arg)
				case pb.MessageID_S2C_FIGHT_BOUT_RESULT:
					r.onBattleBoutResult(call.arg)
				case pb.MessageID_S2C_BATTLE_END:
					r.onBattleEnd(call.arg)
				}

			case <-closeSignal:
				break L1
			}

		}

		wait.Done()

	}()
}

func getRpcHandler(msgID pb.MessageID) network.RpcHandler {
	return func(ses *network.Session, arg interface{}) (reply interface{}, err error) {
		id := ses.GetProp("id")
		if id == nil {
			return
		}

		robotID := id.(uint32)
		r, ok := id2Robot[robotID]
		if !ok {
			return
		}

		r.rpcChan <- &rpcCall{
			msgID: msgID,
			arg:   arg,
		}

		return
	}
}

type robotService struct {
}

func (rs *robotService) Start(appID uint16) {
	closeSignal = make(chan struct{})
	id2Robot = make(map[uint32]*robot)
	peer = network.NewPeer(&network.PeerConfig{
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		MaxPacketSize: 1024 * 1024,
	})

	for _, msgID := range []pb.MessageID{
		pb.MessageID_S2C_MATCH_TIMEOUT,
		pb.MessageID_S2C_UPDATE_MATCH_INFO,
		pb.MessageID_S2C_FIGHT_BOUT_BEGIN,
		pb.MessageID_S2C_FIGHT_BOUT_RESULT,
		pb.MessageID_S2C_BATTLE_END,
	} {

		peer.RegisterRpcHandler(msgID, getRpcHandler(msgID))
	}

	for i := 0; i < robotAmount; i++ {
		id := robotMinID + uint32(i)
		gate := gates[gateIndex%len(gates)]
		gateIndex++
		gateIP := gate[0].(string)
		gatePort := gate[1].(int)
		r := newRobot(id, gateIP, gatePort)
		id2Robot[id] = r
		wait.Add(1)
		r.fuckIt()
	}
}

func (rs *robotService) Stop() {
	close(closeSignal)
	wait.Wait()
	peer.Close()
	evq.Stop()
}

func main() {
	app.NewApplication("robot", &robotService{}).Run()
}
