package main

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
)

func fixBug190104() {
	cityIDs := []int{67, 68, 69, 70, 55, 54, 53, 52, 51, 45, 33}
	uid := common.UUid(15091)
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		glog.Errorf("fixBug190104 no player")
		return
	}

	myCityID := p.getCityID()
	myCity := p.getCity()
	if myCity == nil {
		glog.Errorf("fixBug190104 no city")
		return
	}

	myOldCty := p.getCountry()
	err := myCity.autocephaly(true)
	if err != nil {
		glog.Errorf("fixBug190104 autocephaly %s", err)
		return
	}

	myCty := p.getCountry()
	if myCty == nil || myCty.getID() == myOldCty.getID() {
		glog.Errorf("fixBug190104 fuck country")
		return
	}

	for _, cityID := range cityIDs {
		if cityID == myCityID {
			continue
		}

		cty := cityMgr.getCity(cityID)
		if cty == nil {
			glog.Errorf("fixBug190104 no city %d", cityID)
			continue
		}

		cityCountry := cty.getCountry()
		if cityCountry.getID() != myOldCty.getID() {
			glog.Errorf("fixBug190104 city %d, country %d", cityID, cityCountry.getID())
			continue
		}

		yourMajesty := cityCountry.getYourMajesty()
		if yourMajesty != nil && yourMajesty.getCityID() == cityID {
			glog.Errorf("fixBug190104 yourMajesty in city %d", cityID)
			continue
		}

		cty.surrender( cty.getPrefect(), myCty, true )
	}
}

func fixBug190105() {
	cty := cityMgr.getCity(39)
	p, _ := playerMgr.loadPlayer(9003)
	if p == nil || cty == nil {
		return
	}

	if p.getCityID() != cty.getCityID() {
		return
	}

	cty.modifyResource(resForage, 30)
}
