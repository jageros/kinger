// Generated by gen_errcode.py
// DO NOT EDIT!

package gamedata

import "strconv"

type GameError int32

func (e GameError) Error() string {
	return strconv.Itoa(int(e))
}

func (e GameError) Errcode() int32 {
	return int32(e)
}

const (
	Success      GameError = 0
	InternalErr  GameError = -2
	LevelLockErr GameError = 100
	TauntErr     GameError = 101
)
