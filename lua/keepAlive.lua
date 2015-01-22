-- set the ttl
redis.call("expire", KEYS[1], ARGV[1])