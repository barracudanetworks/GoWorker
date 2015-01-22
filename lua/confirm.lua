-- given a hash wich is provided via KEYS[1]
-- delete the associated temporary job and lock

-- generate the keys to remove
local job = "tmp_job:value:" .. KEYS[1]
local lock = "tmp_job:lock:" .. KEYS[1]

-- remove the keys
return redis.call("del", job,lock)