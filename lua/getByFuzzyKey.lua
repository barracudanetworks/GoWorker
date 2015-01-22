local keys = redis.call("keys", KEYS[1] .. "*")
local vals = redis.call("mget", unpack(keys))
return vals