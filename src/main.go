package main

import (
	"flag"
	"github.com/davy66666/poker-go/src/github.com/dolotech/leaf"
	"github.com/davy66666/poker-go/src/github.com/dolotech/lib/db"
	"github.com/davy66666/poker-go/src/github.com/golang/glog"
	"github.com/davy66666/poker-go/src/server/conf"
	"github.com/davy66666/poker-go/src/server/game"
	"github.com/davy66666/poker-go/src/server/gate"
	"github.com/davy66666/poker-go/src/server/login"
	"github.com/davy66666/poker-go/src/server/model"
	"net/http"
)

var Commit = ""
var BUILD_TIME = ""
var VERSION = ""

var createdb bool

func init() {
	flag.StringVar(&conf.Server.WSAddr, "addr", ":8989", "websocket port")
	flag.IntVar(&conf.Server.MaxConnNum, "maxconn", 20000, "Max Conn Num")
	flag.StringVar(&conf.Server.DBUrl, "pg", "postgres://postgres:haosql@127.0.0.1:5432/postgres?sslmode=disable", "pg addr")
	flag.BoolVar(&createdb, "createdb", false, "initial database")

	flag.Parse()

	db.Init(conf.Server.DBUrl)
	if createdb {
		createDb()
	}
}

func main() {

	go leaf.Run(
		game.Module,
		gate.Module,
		login.Module,
	)

	// for test client
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		glog.Fatalf("ListenAndServe: %v ", err)
	}
}

func createDb() {
	// 建表,只维护和服务器game里面有关的表
	err := db.C().Sync(model.User{}, model.Room{})
	if err != nil {
		glog.Errorln(err)
	}
}
