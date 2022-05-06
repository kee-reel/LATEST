package workers

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"web/models"
	"web/utils"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type Workers struct {
	runner_result_queue redis.Conn
	runner_job_queue    redis.Conn
	jobs                sync.Map
	job_timeout         time.Duration
	task_queue_name     string
	result_queue_name   string
}

func NewWorkers() *Workers {
	job_timeout, err := time.ParseDuration(utils.Env("WEB_JOB_TIMEOUT"))
	utils.Err(err)

	w := Workers{
		utils.CreateRedisConn(),
		utils.CreateRedisConn(),
		sync.Map{},
		job_timeout,
		utils.Env("REDIS_TASK_LIST"),
		utils.Env("REDIS_RESULT_LIST"),
	}
	go w.processJobs()
	return &w
}

func (w *Workers) processJobs() {
	for {
		test_result_json, err := redis.ByteSlices(w.runner_result_queue.Do("BRPOP", w.result_queue_name, 0))
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
	_, err = w.runner_job_queue.Do("RPUSH", w.task_queue_name, runner_data_json)
	utils.Err(err)

	select {
	case res := <-job_ch:
		return res
	case <-time.After(w.job_timeout):
		return nil
	}
}
