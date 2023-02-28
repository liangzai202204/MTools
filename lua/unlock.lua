
if redis.call('get', KEYS[1]) == ARGV[1] then
    --    确实是你的锁
    return redis.call('del', KEYS[1])
else
--    不是你的锁
    return 0
end