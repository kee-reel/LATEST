package limits

import (
	"fmt"
	"log"
	"time"
	"web/utils"

	"github.com/gomodule/redigo/redis"
)

type Limit struct {
	Rate  float32
	Burst float32
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

func (l *Limits) HandleCall(call_type int, client_id string) float32 {
	key := fmt.Sprintf("%d:%s", call_type, client_id)
	data, err := redis.String(l.kv.Do("GET", key))
	lim, ok := l.call_to_lim[call_type]
	if !ok {
		panic(fmt.Sprintf("Limit for call %d not handled", call_type))
	}

	var volume float32
	var old_dt int32
	dt := int32(time.Now().Unix())
	if err == nil {
		_, err := fmt.Sscanf(data, "%f:%d", &volume, &old_dt)
		utils.Err(err)
		dt_diff := dt - old_dt
		volume_diff := float32(dt_diff) * lim.Rate
		if volume_diff > volume {
			volume = 0
		} else {
			volume -= volume_diff
		}
	}
	volume += 1
	if volume <= lim.Burst {
		expire_in := int32(volume * 1000 / lim.Rate)
		data = fmt.Sprintf("%.1f:%d", volume, dt)
		_, err := l.kv.Do("SET", key, data, "PX", expire_in)
		utils.Err(err)
		return 0
	}
	log.Print(volume, lim.Burst, lim.Rate)
	return (volume - lim.Burst) / lim.Rate
}
