package redis

import (
	"log"
	"time"

	"github.com/barracudanetworks/GoWorker/lua"
)

// keepAlive holds a function that keeps the lock alive
type keepAlive struct {
	killChan chan struct{}
	ttl      time.Duration
	key      string
	job      *RedisJob
}

// KeepAlive keep this bitch alive
func (k *keepAlive) KeepAlive(r *Redis) {
	var err error
	for {
		select {
		case <-time.After(k.ttl / 2):
			r.Lock()
			_, err = lua.KEEP_ALIVE_SCRIPT.Do(r.conn, k.key, int(k.ttl.Seconds()))
			if err != nil {
				log.Println(err)
				return
			}
			r.Unlock()

		case <-k.killChan:
			log.Println("Job", k.key, "completed")
			return
		}
	}
}

// Kill stop keeping the lock alive
func (k *keepAlive) Kill() {
	k.killChan <- struct{}{}
}
