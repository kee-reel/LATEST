package workers

import (
	"encoding/json"
	"late/models"
	"late/storage"
	"late/utils"
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type Workers struct {
	kv          redis.Conn
	jobs        sync.Map
	job_timeout time.Duration
}

func NewWorkers() *Workers {
	job_timeout, err := time.ParseDuration(utils.Env("WEB_JOB_TIMEOUT"))
	utils.Err(err)

	w := Workers{
		storage.CreateRedisConn(),
		sync.Map{},
		job_timeout,
	}
	go w.processJobs()
	return &w
}

func (w *Workers) processJobs() {
	for {
		test_result_json, err := redis.ByteSlices(w.kv.Do("BRPOP", utils.Env("REDIS_TESTS_LIST"), 0))
		if err != nil {
			log.Printf("Got redis error while processing jobs: %s, waiting", err)
			time.Sleep(time.Second * 15)
			continue
		}
		if len(test_result_json) != 2 {
			log.Print("List poped more than one element")
			continue
		}

		var test_result models.TestResult
		err = json.Unmarshal(test_result_json[1], &test_result)
		value, ok := w.jobs.LoadAndDelete(test_result.Id)
		if !ok {
			log.Printf("Could not find job with id: %d", test_result.Id)
			continue
		}
		*(value.(*chan *models.TestResult)) <- &test_result
	}
}

func (w *Workers) DoJob(runner_data *models.RunnerData) *models.TestResult {
	runner_data_json, err := json.Marshal(runner_data)
	utils.Err(err)

	job_ch := make(chan *models.TestResult)
	w.jobs.Store(runner_data.Id, &job_ch)
	_, err = w.kv.Do("RPUSH", utils.Env("REDIS_SOLUTIONS_LIST_PREFIX"), runner_data_json)
	utils.Err(err)

	select {
	case res := <-job_ch:
		return res
	case <-time.After(w.job_timeout):
		return nil
	}
}
