package lua

import (
	"io/ioutil"
	"log"

	"github.com/barracudanetworks/GoWorker/config"
	redigo "github.com/garyburd/redigo/redis"
)

var (
	GET_BY_FUZZY_KEY_SCRIPT = func() *redigo.Script {
		b, err := ioutil.ReadFile(config.LUA_PATH + "/getByFuzzyKey.lua")
		if err != nil {
			log.Fatal(err)
		}
		return redigo.NewScript(1, string(b))
	}()
)
