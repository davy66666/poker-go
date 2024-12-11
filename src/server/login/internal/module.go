package internal

import (
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/module"
	"github.com/davy66666/poker-go/src/server/base"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) OnDestroy() {

}
