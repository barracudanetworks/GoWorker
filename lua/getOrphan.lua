local keyMatchPatern = "%w+:%w+:(%w+)"

-- find every orphan key
local function findOrphanKeys(keys,locks)
	local newKeys = {}
	local i = 1
	for k,v in pairs(keys) do
		if locks[k] == nil then
			newKeys[#newKeys + 1] = v
		end
	end
	return newKeys
end

-- for every value-key, check if it has a lock
local function FlipHashes(t)
	local tbl = {}
	for k,v in pairs(t) do 
		local s = string.match(v, keyMatchPatern)
		tbl[s] = v
	end
	return tbl
end

-- get hashes
local function getHashes(t) 
	local keys = {}
	for k, v in pairs(t) do
		local s = string.match(v, keyMatchPatern)
		keys[k] = s 
	end
	return keys
end
-- get all of the keys and locks of all jobs in progress
local keys = redis.call("keys", "tmp_job:value:*")
local locks = redis.call("keys", "tmp_job:lock:*")
local value_keys = keys
-- get all of the hashes from the complex keys
locks = FlipHashes(locks)
keys = FlipHashes(keys)

-- find all of the orphans if they exist 
local orphanKeys = findOrphanKeys(keys,locks)

-- if there are no orphans, just return a empty table
if #orphanKeys == 0 then
	return orphanKeys 
end

-- get all of the orphans and set a lock on them
local orphans =  redis.call("mget", unpack(orphanKeys))

-- get all of the orphan hashes 
local orphanHashes = getHashes(orphanKeys)

local max = #orphanKeys 
if ARGV[1] ~= nil then
	max = tonumber(ARGV[1])
end

local count = 0
-- set locks and ttl's on each of the new jobs
for k,v in pairs(orphanKeys) do
	redis.call("set", "tmp_job:lock:" .. v, "")
	redis.call("expire", "tmp_job:lock:" .. v, 30)
	if count >= max then
		break
	end
	count = count + 1
end

return {unpack(orphans, 1, max)}