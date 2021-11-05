package cardpool

import (
	"math/rand"

	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
)

const (
	minDiyCardID uint32 = 1000000
)

var _ types.IFightCardData = &diyCard{}

type diyCard struct {
	attr *attribute.MapAttr
}

func (dc *diyCard) GetName() string {
	return dc.attr.GetStr("name")
}

func (dc *diyCard) GetGCardID() uint32 {
	return dc.attr.GetUInt32("cardID")
}

func (dc *diyCard) GetCardID() uint32 {
	return dc.attr.GetUInt32("cardID")
}

func (dc *diyCard) GetSkillIds() []int32 {
	gdata := gamedata.GetGameData(consts.Diy).(*gamedata.DiyGameData)
	diyData1 := gdata.GetDiyData(dc.getDiySkillId1())
	diyData2 := gdata.GetDiyData(dc.getDiySkillId2())
	var skills []int32
	if diyData1 != nil {
		skills = append(skills, diyData1.Skill...)
	}
	if diyData2 != nil {
		skills = append(skills, diyData2.Skill...)
	}
	return skills
}

func (dc *diyCard) RandomUp() int {
	return dc.attr.GetInt("minUp") + rand.Intn(dc.attr.GetInt("maxUp")-dc.attr.GetInt("minUp")+1)
}

func (dc *diyCard) RandomDown() int {
	return dc.attr.GetInt("minDown") + rand.Intn(dc.attr.GetInt("maxDown")-dc.attr.GetInt("minDown")+1)
}

func (dc *diyCard) RandomLeft() int {
	return dc.attr.GetInt("minLeft") + rand.Intn(dc.attr.GetInt("maxLeft")-dc.attr.GetInt("minLeft")+1)
}

func (dc *diyCard) RandomRight() int {
	return dc.attr.GetInt("minRight") + rand.Intn(dc.attr.GetInt("maxRight")-dc.attr.GetInt("minRight")+1)
}

func (dc *diyCard) GetUpValueRate() float32 {
	return 0
}

func (dc *diyCard) GetDownValueRate() float32 {
	return 0
}

func (dc *diyCard) GetLeftValueRate() float32 {
	return 0
}

func (dc *diyCard) GetRightValueRate() float32 {
	return 0
}

func (dc *diyCard) GetAdjFValue() float32 {
	return 0
}

func (dc *diyCard) GetCardValue() float32 {
	return 20
}

func (dc *diyCard) GetCamp() int {
	return consts.Heroes
}

func (dc *diyCard) GetLevel() int {
	gdata := gamedata.GetGameData(consts.Diy).(*gamedata.DiyGameData)
	diyData1 := gdata.GetDiyData(dc.getDiySkillId1())
	diyData2 := gdata.GetDiyData(dc.getDiySkillId2())

	lv1 := 0
	if diyData1 != nil {
		lv1 = diyData1.Level
	}

	lv2 := 0
	if diyData2 != nil {
		lv2 = diyData2.Level
	}

	if lv1 > lv2 {
		return lv1
	} else {
		return lv2
	}
}

func (dc *diyCard) GetCardType() int {
	return consts.CtGeneral
}

func (dc *diyCard) getWeapon() string {
	return dc.attr.GetStr("weapon")
}

func (dc *diyCard) PackDiyFightCardInfo() *pb.DiyFightCardInfo {
	return &pb.DiyFightCardInfo{
		CardId:      dc.GetCardID(),
		Name:        dc.GetName(),
		DiySkillId1: int32(dc.getDiySkillId1()),
		DiySkillId2: int32(dc.getDiySkillId2()),
		Weapon:      dc.getWeapon(),
	}
}

func (dc *diyCard) packDataMsg() *pb.DiyCardData {
	return &pb.DiyCardData{
		CardId:      dc.GetCardID(),
		Name:        dc.GetName(),
		DiySkillId1: int32(dc.getDiySkillId1()),
		DiySkillId2: int32(dc.getDiySkillId2()),
		MinUp:       int32(dc.attr.GetInt("minUp")),
		MaxUp:       int32(dc.attr.GetInt("maxUp")),
		MinDown:     int32(dc.attr.GetInt("minDown")),
		MaxDown:     int32(dc.attr.GetInt("maxDown")),
		MinLeft:     int32(dc.attr.GetInt("minLeft")),
		MaxLeft:     int32(dc.attr.GetInt("maxLeft")),
		MinRight:    int32(dc.attr.GetInt("minRight")),
		MaxRight:    int32(dc.attr.GetInt("maxRight")),
		Weapon:      dc.getWeapon(),
	}
}

func (dc *diyCard) packReplyMsg() *pb.DiyCardReply {
	return &pb.DiyCardReply{
		CardId:   dc.GetCardID(),
		MinUp:    int32(dc.attr.GetInt("minUp")),
		MaxUp:    int32(dc.attr.GetInt("maxUp")),
		MinDown:  int32(dc.attr.GetInt("minDown")),
		MaxDown:  int32(dc.attr.GetInt("maxDown")),
		MinLeft:  int32(dc.attr.GetInt("minLeft")),
		MaxLeft:  int32(dc.attr.GetInt("maxLeft")),
		MinRight: int32(dc.attr.GetInt("minRight")),
		MaxRight: int32(dc.attr.GetInt("maxRight")),
	}
}

func (dc *diyCard) getResource() (wine int, book int) {
	wine = dc.attr.GetInt("wine")
	book = dc.attr.GetInt("book")
	return
}

func (dc *diyCard) getDiySkillId1() int {
	return dc.attr.GetInt("diySkillId1")
}

func (dc *diyCard) getDiySkillId2() int {
	return dc.attr.GetInt("diySkillId2")
}

func (dc *diyCard) setCardID(cardID uint32) {
	dc.attr.SetUInt32("cardID", cardID)
}

func (dc *diyCard) setDiySkillId(diySkillId1, diySkillId2 int) {
	dc.attr.SetInt("diySkillId1", diySkillId1)
	dc.attr.SetInt("diySkillId2", diySkillId2)
}

func (dc *diyCard) setResource(wine, book int) {
	dc.attr.SetInt("wine", wine)
	dc.attr.SetInt("book", book)
}

func (dc *diyCard) setName(name string) {
	dc.attr.SetStr("name", name)
}

func (dc *diyCard) setWeapon(weapon string) {
	dc.attr.SetStr("weapon", weapon)
}

func (dc *diyCard) setUpNum(min, max int) {
	dc.attr.SetInt("minUp", min)
	dc.attr.SetInt("maxUp", max)
}

func (dc *diyCard) setDownNum(min, max int) {
	dc.attr.SetInt("minDown", min)
	dc.attr.SetInt("maxDown", max)
}

func (dc *diyCard) setLeftNum(min, max int) {
	dc.attr.SetInt("minLeft", min)
	dc.attr.SetInt("maxLeft", max)
}

func (dc *diyCard) setRightNum(min, max int) {
	dc.attr.SetInt("minRight", min)
	dc.attr.SetInt("maxRight", max)
}

func newDiyCardByAttr(attr *attribute.MapAttr) *diyCard {
	return &diyCard{
		attr: attr,
	}
}
