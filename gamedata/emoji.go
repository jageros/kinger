package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type Emoji struct {
	ID string `json:"__id__"`
	Team int `json:"team"`
	TeamName int `json:"teamName"`
}

func (e *Emoji) GetTeamName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(e.TeamName)
}

type EmojiGameData struct {
	baseGameData
	EmojiTeams []*Emoji
	Team2Emoji map[int]*Emoji
}

func newEmojiGameData() *EmojiGameData {
	gd := &EmojiGameData{}
	gd.i = gd
	return gd
}

func (gd *EmojiGameData) name() string {
	return consts.Emoji
}

func (gd *EmojiGameData) init(d []byte) error {
	var l []*Emoji
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.EmojiTeams = []*Emoji{}
	gd.Team2Emoji = map[int]*Emoji{}
	for _, e := range l {
		if _, ok := gd.Team2Emoji[e.Team]; !ok {
			gd.Team2Emoji[e.Team] = e
			gd.EmojiTeams = append(gd.EmojiTeams, e)
		}
	}

	return nil
}
