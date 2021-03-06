package lua

import (
	"io/ioutil"
	"log"

	redigo "github.com/garyburd/redigo/redis"

	"github.com/barracudanetworks/GoWorker/config"
)

var (
	// pop and lock pops a key off of a list, puts it in a temporary key, and sets a lock
	// ARGS: 0 list key 1 ttl = 30 seconds if not given
	GET_ORPHAN_SCRIPT = func() *redigo.Script {
		b, err := ioutil.ReadFile(config.LUA_PATH + "/getOrphan.lua")
		if err != nil {
			log.Fatal(err)
		}
		return redigo.NewScript(0, string(b))
	}()
)
