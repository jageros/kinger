package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type ChatPop struct {
	ID string `json:"__id__"`
}

type ChatPopGameData struct {
	baseGameData
	ChatPops   []*ChatPop
	ID2ChatPop map[string]*ChatPop
}

func newChatPopGameData() *ChatPopGameData {
	r := &ChatPopGameData{}
	r.i = r
	return r
}

func (c *ChatPopGameData) name() string {
	return consts.ChatPopConfig
}

func (c *ChatPopGameData) init(d []byte) error {
	var l []*ChatPop
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	c.ChatPops = l
	c.ID2ChatPop = map[string]*ChatPop{}
	for _, h := range l {
		c.ID2ChatPop[h.ID] = h
	}
	return nil
}
