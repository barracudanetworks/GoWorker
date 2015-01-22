
-- grab the value
local val = redis.call("lpop", KEYS[1])

local hash = redis.sha1hex(val)
-- init keys
local valKey = "tmp_job:value:" .. hash
local lockKey = "tmp_job:lock:" .. hash

-- set it as a temparay job 
redis.call("set", valKey, val)

-- set a lock on the file
redis.call("set", lockKey,"")

-- set a ttl on lock
local ttl = 30
if type(ARGV[1]) ~= nil then
	ttl = ARGV[1]
end
redis.call("expire", lockKey, ttl)

-- return the value and lockKey
return val 