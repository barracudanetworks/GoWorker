package lua

import (
	"io/ioutil"
	"log"

	"github.com/barracudanetworks/GoWorker/config"
	redigo "github.com/garyburd/redigo/redis"
)

var (
	CONFIRM_SCRIPT = func() *redigo.Script {
		b, err := ioutil.ReadFile(config.LUA_PATH + "/confirm.lua")
		if err != nil {
			log.Fatal(err)
		}
		return redigo.NewScript(1, string(b))
	}()
)
