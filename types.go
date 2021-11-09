package agent

import (
	"sync"

	ndk "github.com/nokia/srlinux-ndk-go/v21/ndk"
)

type BfdSession struct {
	m     *sync.RWMutex
	BySrc map[uint32]map[string]map[string][]*ndk.BfdmgrGeneralSessionDataPb // instance / srcIP / dstIP
	ByDst map[uint32]map[string]map[string][]*ndk.BfdmgrGeneralSessionDataPb // instance / dstIP / srcIP
}

func newBfdSession() *BfdSession {
	return &BfdSession{
		m:     new(sync.RWMutex),
		BySrc: make(map[uint32]map[string]map[string][]*ndk.BfdmgrGeneralSessionDataPb),
		ByDst: make(map[uint32]map[string]map[string][]*ndk.BfdmgrGeneralSessionDataPb),
	}
}
