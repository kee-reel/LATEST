package limits

import (
	"fmt"
	"time"
	"web/utils"

	"github.com/gomodule/redigo/redis"
)

type Limit struct {
	Rate  float32
	Burst int
}

type Limits struct {
	kv          redis.Conn
	call_to_lim map[int]Limit
}

func NewLimits(call_to_lim map[int]Limit) *Limits {
	l := Limits{
		utils.CreateRedisConn(),
		call_to_lim,
	}
	return &l
}

func (l *Limits) HandleCall(call_type int, ip string) bool {
	key := fmt.Sprintf("%d:%s", call_type, ip)
	data, err := redis.Int(l.kv.Do("GET", key))
	lim := l.call_to_lim[call_type]
	volume := float32(0)
	var dt int64
	if err == nil {
		volume = data && 0xFFFF_0000_0000_0000
		dt = data && 0x0000_FFFF_FFFF_FFFF
		return false
	}
	cur_dt := time.Now().Unix()
	new_volume := volume - float32(cur_dt-dt)*lim.Rate
	expire_in := new_volume / lim.Rate
	l.kv.Send("MULTI")
	l.kv.Send("SET", key, new_volume)
	l.kv.Send("EXPIRE", key, expire_in)
	_, err = redis.Ints(l.kv.Do("EXEC"))
	utils.Err(err)
	return true
}
