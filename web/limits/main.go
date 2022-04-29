package limits

import (
	"fmt"
	"late/utils"

	"github.com/gomodule/redigo/redis"
)

type Limit struct {
	Calls   int
	Timeout int
}

type Limits struct {
	kv          redis.Conn
	call_to_lim map[int]Limit
	default_lim Limit
}

func NewLimits(call_to_lim map[int]Limit, default_lim Limit) *Limits {
	l := Limits{
		utils.CreateRedisConn(),
		call_to_lim,
		default_lim,
	}
	return &l
}

func (l *Limits) HandleCall(call_type int, ip string) bool {
	key := fmt.Sprintf("%d:%s", call_type, ip)
	value, err := redis.Int(l.kv.Do("GET", key))
	var lim Limit
	if err == nil {
		var ok bool
		lim, ok = l.call_to_lim[call_type]
		if !ok {
			lim = l.default_lim
		}
		if value > lim.Calls {
			return false
		}
	}
	l.kv.Send("MULTI")
	l.kv.Send("INCR", key)
	l.kv.Send("EXPIRE", key, lim.Timeout)
	_, err = redis.Ints(l.kv.Do("EXEC"))
	utils.Err(err)
	return true
}
