package network

import "time"

type PeerConfig struct {
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	MaxPacketSize uint32
	ReadBufSize   int
	WriteBufSize  int
}
