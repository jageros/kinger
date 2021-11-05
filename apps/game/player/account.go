package player

import (
	"kinger/gopuppy/attribute"
	"time"
	"fmt"
	"math/rand"
	"kinger/apps/game/module"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"strconv"
	"kinger/gamedata"
	"crypto/md5"
	"io"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common"
	"errors"
	"kinger/gopuppy/apps/logic"
)

type accountSt struct {
	accountID string
	attr *attribute.AttrMgr
	isNew bool
}

func genAccountID(channel, loginChannel, channelID string, isTourist bool) string {
	accountID := fmt.Sprintf("%s_%s", channel, channelID)
	if loginChannel != "" && loginChannel != channel {
		accountID += "_" + loginChannel
	}
	if isTourist {
		accountID += "_tourist"
	}
	return accountID
}

func loadAccountByAccountID(accountID string, regionArg ...uint32) (*accountSt, error) {
	var region uint32
	if len(regionArg) > 0 {
		region = regionArg[0]
	}
	accountAttr := attribute.NewAttrMgr("account", accountID, false, region)
	err := accountAttr.Load()
	return &accountSt{
		accountID: accountID,
		attr: accountAttr,
		isNew: err != nil,
	}, err
}

func loadAccount(channel, loginChannel, channelID string, isTourist bool) (*accountSt, error) {
	return loadAccountByAccountID(genAccountID(channel, loginChannel, channelID, isTourist))
}

func md5HashPassword(password string) string {
	md5Writer := md5.New()
	io.WriteString(md5Writer, password+pwdSuffix)
	return fmt.Sprintf("%x", md5Writer.Sum(nil))
}

func doRegistAccount(channel , account, password string) (*accountSt, error) {
	a, err := loadAccount(channel, "", account, false)
	if err == nil {
		return nil, gamedata.GameError(1)
	}

	if err != attribute.NotExistsErr {
		return nil, err
	}

	a.attr.SetBool("isCpAccount", true)
	a.setPwd(md5HashPassword(password))
	if err = a.attr.Insert(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *accountSt) getArea() int {
	return a.attr.GetInt("area")
}

func (a *accountSt) isCpAccount() bool {
	return a.attr.GetBool("isCpAccount")
}

func (a *accountSt) setWxOpenID(openID string) {
	a.attr.SetStr("wxOpenID", openID)
	a.save(false)
}

func (a *accountSt) getArchive(archiveID int) *attribute.MapAttr {
	return a.attr.GetMapAttr(strconv.Itoa(archiveID))
}

func (a *accountSt) getUid() common.UUid {
	arc := a.getArchive(1)
	if arc == nil {
		return 0
	}
	return common.UUid(arc.GetUInt64("uid"))
}

func (a *accountSt) newArchiveByUid(archiveID int, uid common.UUid) error {
	arc := attribute.NewMapAttr()
	arc.SetInt("id", archiveID)
	arc.SetUInt64("uid", uint64(uid))
	arc.SetUInt32("lastTime", uint32(time.Now().Unix()))
	a.attr.SetMapAttr(strconv.Itoa(archiveID), arc)
	return a.save(true)
}

func (a *accountSt) genArchive(archiveID int) (common.UUid, error) {

	id, err := common.Gen32UUid("player")
	if err != nil {
		glog.Errorf("Gen32UUid player err %s", err)
		return 0, err
	}

	uid := common.UUid(id + 8964)
	if err := a.newArchiveByUid(archiveID, uid); err != nil {
		return 0, err
	}

	return uid, nil
}

func (a *accountSt) setArchive(archiveID int, arc *attribute.MapAttr) error {
	a.attr.SetMapAttr(strconv.Itoa(archiveID), arc)
	return a.save(true)
}

func (a *accountSt) setRawPwd(pwd string) {
	a.attr.SetStr("rawPwd", pwd)
}

func (a *accountSt) getRawPwd() string {
	return a.attr.GetStr("rawPwd")
}

func (a *accountSt) setPwd(pwd string) {
	a.attr.SetStr("pwd", pwd)
}

func (a *accountSt) getPwd() string {
	return a.attr.GetStr("pwd")
}

func (a *accountSt) getAccountID() string {
	return a.accountID
}

func (a *accountSt) save(needReply bool) error {
	if a.isNew {
		a.isNew = false
		a.attr.SetDirty(true)
	}
	return a.attr.Save(needReply)
}

func (a *accountSt) packMsg() *pb.AccountArchives {
	reply := &pb.AccountArchives{Ok: true}
	a.attr.ForEachKey(func(key string) {
		if _, err := strconv.Atoi(key); err != nil {
			return
		}

		archive := a.attr.GetMapAttr(key)
		reply.Archives = append(reply.Archives, &pb.Archive{
			ID:       int32(archive.GetInt("id")),
			Uid:      archive.GetUInt64("uid"),
			LastTime: archive.GetUInt32("lastTime"),
		})
	})
	return reply
}

func (a *accountSt) backupArc() error {
	arc := a.attr.GetMapAttr("1")
	if arc == nil {
		return nil
	}
	a.attr.Del("1")
	arcBak := attribute.NewMapAttr()
	arcBak.AssignMap( arc.ToMap() )
	a.attr.SetMapAttr("1bak", arcBak)
	return a.save(true)
}

func (a *accountSt) getRegion() uint32 {
	arc := a.getArchive(1)
	if arc != nil {
		return logic.GetAgentRegion(common.UUid(arc.GetUInt64("uid")))
	} else {
		return 0
	}
}

func (a *accountSt) bindOthAccount(oth *accountSt) error {
	arc := oth.getArchive(1)
	if arc == nil {
		return errors.New("no arc")
	}

	err := a.backupArc()
	if err != nil {
		return err
	}

	newArc := attribute.NewMapAttr()
	newArc.AssignMap( arc.ToMap() )
	err = a.setArchive(1, newArc)
	if err != nil {
		return err
	}

	othUid := common.UUid(newArc.GetUInt64("uid"))
	accountID := a.getAccountID()
	othRegion := logic.GetAgentRegion(othUid)
	if othRegion != module.Service.GetRegion() {
		newAccountAttr := attribute.NewAttrMgr("account", accountID, false, othRegion)
		newAccountAttr.AssignMap( a.attr.ToMap() )
		if err := newAccountAttr.Insert(); err != nil {
			return err
		}

		mod.onBindAccount(accountID, othRegion)
		logic.BroadcastBackend(pb.MessageID_G2G_ON_BIND_ACCOUNT, &pb.OnBindAccountArg{
			AccountID:  accountID,
			BindRegion: othRegion,
		})

		attr := attribute.NewAttrMgr("bindAccountRegion", accountID, true)
		attr.SetUInt32("region", othRegion)
		attr.Save(true)
	}

	err = oth.backupArc()
	if err != nil {
		return err
	}

	return nil
}

type accountCodeSt struct {
	attr *attribute.AttrMgr
}

func newAccountCode(attr *attribute.AttrMgr) *accountCodeSt {
	return &accountCodeSt{attr: attr}
}

func loadAccountCode(code string) *accountCodeSt {
	attr := attribute.NewAttrMgr("account_code", code, true)
	err := attr.Load()
	if err != nil {
		return nil
	}
	acCode := &accountCodeSt{
		attr: attr,
	}

	if acCode.isTimeout() {
		acCode.del(false)
		return nil
	}
	return acCode
}

func genAccountCode(p *Player) *accountCodeSt {
	for tryCnt := 0; tryCnt < 5; tryCnt++ {
		code := fmt.Sprintf("%d%d%d%d%d%d%d%d", rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10),
			rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))
		acCode := loadAccountCode(code)
		if acCode != nil {
			continue
		}

		attr := attribute.NewAttrMgr("account_code", code, true)
		attr.SetUInt64("uid", uint64(p.GetUid()))
		timeout := time.Now().Unix() + 300
		attr.SetInt64("timeout", timeout)
		attr.SetInt("pvpLevel", p.GetPvpLevel())
		attr.SetStr("accountID", p.getAccountID())
		err := attr.Insert()
		if err != nil {
			continue
		}

		oldCode := p.getAccountCode()
		if oldCode != nil {
			oldCode.del(false)
		}

		p.setAccountCode(code)
		return &accountCodeSt{attr: attr}
	}
	return nil
}

func checkAccountCode() {
	if module.Service.GetAppID() != 1 {
		return
	}

	timer.AddTicker(time.Minute, func() {
		attribute.ForEach("account_code", func(attr *attribute.AttrMgr) {
			code := &accountCodeSt{attr: attr}
			if code.isTimeout() {
				code.del(false)
			}
		})
	})
}

func (ac *accountCodeSt) getCode() string {
	return ac.attr.GetAttrID().(string)
}

func (ac *accountCodeSt) isTimeout() bool {
	return ac.attr.GetInt64("timeout") <= time.Now().Unix()
}

func (ac *accountCodeSt) del(needReply bool) error {
	return ac.attr.Delete(needReply)
}

func (ac *accountCodeSt) getPvpLevel() int {
	return ac.attr.GetInt("pvpLevel")
}

func (ac *accountCodeSt) getUid() common.UUid {
	return common.UUid(ac.attr.GetUInt64("uid"))
}

func (ac *accountCodeSt) loadPlayer() *pb.SimplePlayerInfo {
	return mod.GetSimplePlayerInfo(ac.getUid())
}

func (ac *accountCodeSt) setAccountID(accountID string) {
	ac.attr.SetStr("accountID", accountID)
}

func (ac *accountCodeSt) getAccountID() string {
	return ac.attr.GetStr("accountID")
}

func (ac *accountCodeSt) getRegion() uint32 {
	return logic.GetAgentRegion(ac.getUid())
}
