package base

import (
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/chanrpc"
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/module"
	"github.com/davy66666/poker-go/src/server/conf"
)

func NewSkeleton() *module.Skeleton {
	skeleton := &module.Skeleton{
		GoLen:              conf.GoLen,
		TimerDispatcherLen: conf.TimerDispatcherLen,
		AsynCallLen:        conf.AsynCallLen,
		ChanRPCServer:      chanrpc.NewServer(conf.ChanRPCLen),
	}
	skeleton.Init()
	return skeleton
}
