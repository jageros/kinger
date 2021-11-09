package aicardpool

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	curAppName     string
	curAppID       uint32
	pvpLevel2Pools map[int]*cardPool
	loading        chan struct{}
)

type cardPool struct {
	attr      *attribute.AttrMgr
	poolsAttr *attribute.ListAttr
	pvpLevel  int
	pools     [][]uint32
}

func newCardPool(pvpLevel int) *cardPool {
	if curAppName == "" || curAppID <= 0 {
		return nil
	}
	attr := attribute.NewAttrMgr(fmt.Sprintf("aicardpool_%s_%d", curAppName, curAppID), pvpLevel)
	return newCardPoolByAttr(attr)
}

func newCardPoolByAttr(attr *attribute.AttrMgr) *cardPool {
	pvpLevel, ok := attr.GetAttrID().(int)
	if !ok {
		return nil
	}

	p := &cardPool{attr: attr, pvpLevel: pvpLevel}
	poolsAttr := attr.GetListAttr("pools")
	if poolsAttr == nil {
		poolsAttr = attribute.NewListAttr()
		attr.SetListAttr("pools", poolsAttr)
	}
	p.poolsAttr = poolsAttr

	poolsAttr.ForEachIndex(func(index int) bool {
		var pool []uint32
		cardIDsInfo := strings.Split(poolsAttr.GetStr(index), "_")
		for i, strCardID := range cardIDsInfo {
			cardID, _ := strconv.Atoi(strCardID)
			pool = append(pool, uint32(cardID))
			if i == len(cardIDsInfo)-2 {
				sort.Slice(pool, func(i, j int) bool {
					return pool[i] < pool[j]
				})
			}
		}
		p.pools = append(p.pools, pool)
		return true
	})

	return p
}

func (cp *cardPool) getCurIdx() int {
	return cp.attr.GetInt("curIdx")
}
func (cp *cardPool) setCurIdx(curIdx int) {
	cp.attr.SetInt("curIdx", curIdx)
}

func (cp *cardPool) addPool(camp int, cardIDs []uint32) {
	if len(cardIDs) != 5 {
		return
	}
	sort.Slice(cardIDs, func(i, j int) bool {
		return cardIDs[i] < cardIDs[j]
	})
	cardIDs = append(cardIDs, uint32(camp))

	strCardIDs := fmt.Sprintf("%d_%d_%d_%d_%d_%d", cardIDs[0], cardIDs[1], cardIDs[2], cardIDs[3], cardIDs[4], camp)
	if len(cp.pools) < 100 {
		cp.pools = append(cp.pools, cardIDs)
		cp.poolsAttr.AppendStr(strCardIDs)
	} else {
		curIdx := cp.getCurIdx() + 1
		if curIdx >= len(cp.pools) {
			curIdx = 0
		}
		cp.pools[curIdx] = cardIDs
		cp.poolsAttr.SetStr(curIdx, strCardIDs)
	}
}

func (cp *cardPool) save() {
	cp.attr.Save(false)
}

func Load(appName string, appID uint32) {
	curAppName = appName
	curAppID = appID
	loading := make(chan struct{})
	defer func() {
		close(loading)
		loading = nil
	}()

	evq.CallLater(func() {

		level2Pools := map[int]*cardPool{}
		allAttrs, _ := attribute.LoadAll(fmt.Sprintf("aicardpool_%s_%d", appName, appID))
		for _, attr := range allAttrs {
			cp := newCardPoolByAttr(attr)
			if cp == nil {
				continue
			}
			level2Pools[cp.pvpLevel] = cp
		}
		pvpLevel2Pools = level2Pools
	})

	timer.AddTicker(12*time.Minute, Save)
}

func Save() {
	if pvpLevel2Pools == nil {
		return
	}

	for _, cp := range pvpLevel2Pools {
		cp.save()
	}
}

func AddCardPool(pvpLevel, camp int, cardIDs []uint32) {
	if loading != nil {
		evq.Await(func() {
			<-loading
		})
	}

	if pvpLevel2Pools == nil {
		return
	}

	cp, ok := pvpLevel2Pools[pvpLevel]
	if !ok {
		cp = newCardPool(pvpLevel)
		if cp == nil {
			return
		}
		pvpLevel2Pools[pvpLevel] = cp
	}

	cp.addPool(camp, cardIDs)
}

func randomCardPoolByData(pvpLevel int) (int, []uint32) {
	camp := []int{consts.Wei, consts.Shu, consts.Wu}[rand.Intn(3)]
	cardDatas := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).RandomPvpRobotCards(pvpLevel, camp, true)
	cardIDs := make([]uint32, len(cardDatas))
	for i, cardData := range cardDatas {
		cardIDs[i] = cardData.CardID
	}
	return camp, cardIDs
}

func RandomCardPool(pvpLevel, playerCamp int, playerHandCards []*pb.SkinGCard, canEqualHand bool) (int, []*pb.SkinGCard) {
	var handCards []*pb.SkinGCard
	if len(playerHandCards) != 5 {
		return playerCamp, playerHandCards
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	playerHandCardIDs := make([]uint32, len(playerHandCards))
	for i, card := range playerHandCards {
		cardData := poolGameData.GetCardByGid(card.GCardID)
		if cardData != nil {
			playerHandCardIDs[i] = cardData.CardID
		}
	}

	sort.Slice(playerHandCardIDs, func(i, j int) bool {
		return playerHandCardIDs[i] < playerHandCardIDs[j]
	})

	var cardIDs []uint32
	var camp int
	if pvpLevel2Pools == nil {
		camp, cardIDs = randomCardPoolByData(pvpLevel)
	} else {

		pools := []*cardPool{pvpLevel2Pools[pvpLevel-1], pvpLevel2Pools[pvpLevel], pvpLevel2Pools[pvpLevel+1]}
		poolAmounts := []int{0, 0, 0}
		var totalPoolAmount int
		for i, p := range pools {
			if p != nil {
				totalPoolAmount += len(p.pools)
				poolAmounts[i] = len(p.pools)
			}
		}

		if totalPoolAmount <= 0 {
			camp, cardIDs = randomCardPoolByData(pvpLevel)
		} else {

			r := rand.Intn(totalPoolAmount)
			for i, amount := range poolAmounts {
				if amount <= r {
					r -= amount
				} else {
					oldRand := r
					poolAmount := len(pools[i].pools)
					for {
						fromCardIDs := pools[i].pools[r][:5]
						camp = int(pools[i].pools[r][5])
						if !canEqualHand {
							isEqual := true
							for idx, cardID := range playerHandCardIDs {
								if fromCardIDs[idx] != cardID {
									isEqual = false
									break
								}
							}

							if isEqual {
								r = (r + 1) % poolAmount
								if r != oldRand {
									continue
								}
							}
						}

						cardIDs = make([]uint32, len(fromCardIDs))
						copy(cardIDs, fromCardIDs)
						utils.ShuffleUInt32(cardIDs)
						break
					}
					break
				}
			}
		}
	}

	if len(cardIDs) != 5 {
		return playerCamp, playerHandCards
	}

	levels := []int{0, 0, 0, 0, 0}
	//equips := []string{"", "", "", "", ""}
	//var equipAmount int
	for i, card := range playerHandCards {
		//if card.Equip != "" {
		//	equipAmount += 1
		//}

		cardData := poolGameData.GetCardByGid(card.GCardID)
		if cardData == nil {
			levels[i] = 1
		} else if cardData.IsSpCard() {
			levels[i] = 3
		} else {
			levels[i] = cardData.Level
		}
	}

	utils.ShuffleInt(levels)
	/*
		if equipAmount > 0 {
			equipGameData := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
			equipIDs := utils.RandSample(equipGameData.AllEquipIDs, equipAmount, false)
			for i, equipID := range equipIDs {
				equips[i] = equipID.(string)
			}
			utils.ShuffleString(equips)
		}
	*/

	handCards = make([]*pb.SkinGCard, 5)
	for i, cardID := range cardIDs {
		cardData := poolGameData.GetCard(cardID, levels[i])
		if cardData == nil {
			cardData = poolGameData.GetCard(cardID, 1)
			if cardData == nil {
				return playerCamp, playerHandCards
			}
		}
		handCards[i] = &pb.SkinGCard{GCardID: cardData.GCardID}
	}
	return camp, handCards
}
