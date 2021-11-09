package consts

const (
	// evq
	EvNewCard = 1000 + iota
	EvModifyName
	EvReloadConfig
	EvLogin
)

const (
	// eventhub
	EvBeginBattle = 1000 + iota
	EvResUpdate   // 资源变化时，(player types.IPlayer, resType, curAmount, modifyAmount int)
	EvMaxPvpLevelUpdate
	EvPvpLevelUpdate
	EvEndPvpBattle // pvp战斗结束时，(player types.IPlayer, isWin bool, fighterData *pb.EndFighterData)
	EvFixServer1Data
	EvEquipDel
	EvRecharge // 充值时, (player types.IPlayer, rmbPrice int, goodsID string)
	EvReborn
	EvDelOutStatus // 外部状态失效时，(types.IPlayer, types.IOutStatus)
	EvFixServer1WxInvite
	EvFixLevelRechargeUnlock
	EvOpenTreasure      // 开宝箱
	EvCombat            //约战
	EvGetMissionReward  //完成任务
	EvShareBattleReport //分享战报
	EVWatchBattleReport //观看战报
	EvAddFriend         //增加好友
	EvCardUpdate        //卡信息更新
	EvShare             //分享游戏
)
