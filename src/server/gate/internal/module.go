package internal

import (
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/gate"
	"github.com/davy66666/poker-go/src/github.com/golang/glog"
	"github.com/davy66666/poker-go/src/server/conf"
	"github.com/davy66666/poker-go/src/server/game"
	"github.com/davy66666/poker-go/src/server/protocol"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.Server.MaxConnNum,
		PendingWriteNum: conf.PendingWriteNum,
		MaxMsgLen:       conf.MaxMsgLen,
		WSAddr:          conf.Server.WSAddr,
		HTTPTimeout:     conf.HTTPTimeout,
		CertFile:        conf.Server.CertFile,
		KeyFile:         conf.Server.KeyFile,
		TCPAddr:         conf.Server.TCPAddr,
		LenMsgLen:       conf.LenMsgLen,
		LittleEndian:    conf.LittleEndian,
		Processor:       protocol.Processor,
		AgentChanRPC:    game.ChanRPC,
	}
}
func (gate *Module) OnDestroy() {
	glog.Errorln("OnDestroy")
}
