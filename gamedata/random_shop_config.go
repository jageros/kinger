package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/consts"
	"regexp"
	"strconv"
	"strings"
)

type RandomShopConfig struct {
	RequireID 	string 		`json:"__id__"`
	ParaValue 	[]int 	`json:"para_value"`
}

type RandomShopGameData struct {
	baseGameData
	rawData []byte
	GoodsNum int
	RefreshCnt int
	RefreshCd int
	RefreshFree int
	RefreshPrice int
	CardLevelupNum int
	CardLessNum int
	RandomData map[string][]int
	CardPro map[string][]int
	FreeData map[string][]int
	CardStarPrice map[int][]int
	CardStarPriceGold map[int][]int
	GoldPrice []int
	TeamCard map[int][]int
	TeamGold map[int][]int
	TeamFreeJade map[int][]int
	TeamFreeGold map[int][]int
	RandomParaValue int
	FreeRandomPara int
	CardProValue int
}

func (rs *RandomShopGameData) name() string {
	return consts.RandomShop
}

func newRandomShopGameData() *RandomShopGameData {
	gd := &RandomShopGameData{}
	gd.i = gd
	return gd
}


func (rs *RandomShopGameData) init(d []byte) error{
	if bytes.Equal(rs.rawData, d) {
		return errors.New("no update")
	}

	var l []*RandomShopConfig
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	rs.RandomData = map[string][]int{}
	rs.CardPro = map[string][]int{}
	rs.FreeData = map[string][]int{}
	rs.CardStarPrice = map[int][]int{}
	rs.CardStarPriceGold = map[int][]int{}
	rs.TeamCard = map[int][]int{}
	rs.TeamGold = map[int][]int{}
	rs.TeamFreeJade = map[int][]int{}
	rs.TeamFreeGold = map[int][]int{}

	rs.rawData = d
	randomParaValue := 1
	FreeRandomPara := 1
	CardProValue := 1
	for _, d := range l {
		switch d.RequireID {
		case "goods_num":
			rs.GoodsNum = d.ParaValue[0]
		case "refresh_cnt":
			rs.RefreshCnt = d.ParaValue[0]
		case "refresh_cd":
			rs.RefreshCd = d.ParaValue[0]
		case "refresh_free":
			rs.RefreshFree = d.ParaValue[0]
		case "refresh_price":
			rs.RefreshPrice = d.ParaValue[0]
		case "gold_price":
			rs.GoldPrice = d.ParaValue
		case "card_levelup_num":
			rs.CardLevelupNum = d.ParaValue[0]
		case "card_less_num":
			rs.CardLessNum = d.ParaValue[0]
		}

		if strings.HasPrefix(d.RequireID, "random"){
			rs.RandomData[d.RequireID] = []int{randomParaValue, d.ParaValue[0]+randomParaValue-1}
			randomParaValue += d.ParaValue[0]
		}

		if strings.HasPrefix(d.RequireID, "free"){
			rs.FreeData[d.RequireID] = []int{FreeRandomPara, d.ParaValue[0]+FreeRandomPara-1}
			FreeRandomPara += d.ParaValue[0]
		}

		if strings.HasPrefix(d.RequireID, "card") && strings.HasSuffix(d.RequireID,"pro"){
			rs.CardPro[d.RequireID] = []int{CardProValue, d.ParaValue[0]+CardProValue-1}
			CardProValue += d.ParaValue[0]
		}

		if strings.HasPrefix(d.RequireID, "card_star") && strings.HasSuffix(d.RequireID, "price"){
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.CardStarPrice[key] = d.ParaValue
		}

		if strings.HasPrefix(d.RequireID, "card_star") && strings.HasSuffix(d.RequireID, "price_gold"){
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.CardStarPriceGold[key] = d.ParaValue
		}

		if strings.HasPrefix(d.RequireID, "team") && strings.HasSuffix(d.RequireID, "free_jade"){
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.TeamFreeJade[key] = d.ParaValue
		}

		if strings.HasPrefix(d.RequireID, "team") && strings.HasSuffix(d.RequireID, "free_gold"){
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.TeamFreeGold[key] = d.ParaValue
		}

		if strings.HasPrefix(d.RequireID, "team") && strings.HasSuffix(d.RequireID, "card") &&
			strings.Index(d.RequireID, "free") == -1{
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.TeamCard[key] = d.ParaValue
		}

		if strings.HasPrefix(d.RequireID, "team") && strings.HasSuffix(d.RequireID, "gold") &&
			strings.Index(d.RequireID, "free") == -1{
			rep := regexp.MustCompile(`\d+`)
			key, _ := strconv.Atoi(rep.FindAllString(d.RequireID, -1)[0])
			rs.TeamGold[key] = d.ParaValue
		}

	}

	rs.RandomParaValue = randomParaValue
	rs.FreeRandomPara = FreeRandomPara
	rs.CardProValue = CardProValue

	return nil
}
