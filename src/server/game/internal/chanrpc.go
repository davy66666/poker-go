package internal

import (
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/gate"
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/room"
	"github.com/davy66666/poker-go/src/github.com/golang/glog"
	"github.com/davy66666/poker-go/src/server/model"
)

func init() {
	skeleton.RegisterChanRPC(model.Agent_New, rpcNewAgent)
	skeleton.RegisterChanRPC(model.Agent_Close, rpcCloseAgent)
	skeleton.RegisterChanRPC(model.Agent_Login, rpcLoginAgent)
}

func rpcNewAgent(a gate.Agent) {
	glog.Errorln("新建链接 ", a)
}

func rpcCloseAgent(a gate.Agent) {
	glog.Errorln("链接关闭 ", a)
}

func rpcLoginAgent(u *model.User, a gate.Agent) {

	o := NewOccupant(u, a)
	a.SetUserData(o)

	if len(u.RoomID) > 0 {
		o.room = room.GetRoom(u.RoomID)
	}
	glog.Errorln("rpcLoginAgent", u)
}
