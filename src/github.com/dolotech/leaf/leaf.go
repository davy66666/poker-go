package leaf

import (
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf/module"
	"github.com/davy66666/poker-go/src/github.com/golang/glog"
	"os"
	"os/signal"
)

func Run(mods ...module.Module) {

	glog.Errorf("Leaf %v starting up", version)

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	glog.Errorf("Leaf closing down (signal: %v)", sig)
	module.Destroy()
}
