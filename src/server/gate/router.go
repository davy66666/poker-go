package gate

import (
	"github.com/davy66666/poker-go/src/server/game"
	"github.com/davy66666/poker-go/src/server/login"
	"github.com/davy66666/poker-go/src/server/protocol"
)

func init() {
	protocol.Processor.SetRouter(&protocol.UserLoginInfo{}, login.ChanRPC)
	protocol.Processor.SetRouter(&protocol.Version{}, login.ChanRPC)
	protocol.Processor.SetRouter(&protocol.RoomList{}, login.ChanRPC)

	protocol.Processor.SetRouter(&protocol.JoinRoom{}, game.ChanRPC)
	protocol.Processor.SetRouter(&protocol.LeaveRoom{}, game.ChanRPC)
	protocol.Processor.SetRouter(&protocol.SitDown{}, game.ChanRPC)
	protocol.Processor.SetRouter(&protocol.StandUp{}, game.ChanRPC)
	protocol.Processor.SetRouter(&protocol.Bet{}, game.ChanRPC)
	protocol.Processor.SetRouter(&protocol.Chat{}, game.ChanRPC)
}
