package limits

import (
	"fmt"
	"time"
	"web/utils"

	"github.com/gomodule/redigo/redis"
)

type Limit struct {
	Rate  float32 `json:"rate" example:"0.5"`
	Burst float32 `json:"burst" example:"10"`
}

type Limits struct {
	kv redis.Conn
}

func NewLimits() *Limits {
	return &Limits{utils.CreateRedisConn()}
}

func (l *Limits) HandleCall(call_id int, client_id string, lim *Limit) float32 {
	key := fmt.Sprintf("%d:%s", call_id, client_id)
	data, err := redis.String(l.kv.Do("GET", key))

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
	return (volume - lim.Burst) / lim.Rate
}
