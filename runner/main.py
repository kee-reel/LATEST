#!/usr/env python3
import os
import sys
import json
import time
import random
import logging
logging.basicConfig(level=logging.INFO)

import redis

from test import test_solution, PATH
from build import build_solution, LANGS
from schemas import Task, Result

def save_file(solution, field):
    data = solution[field]

    text = data['text']
    extention = data['extention']
    if not extention or not text:
        return None

    fn = os.path.join(PATH, f'{field}.{extention}')

    if text:
        open(fn, 'w').write(text)
    else:
        file_.save(fn)

    return fn


def run_tests(sol_fn, comp_sol_fn, tests):
    sol_fn_new, err = build_solution(sol_fn)
    if sol_fn != sol_fn_new:
        os.remove(sol_fn)
    if err:
        err['tests_passed'] = 0
        return err

    comp_sol_fn_new, err = build_solution(comp_sol_fn)
    if comp_sol_fn != comp_sol_fn_new:
        os.remove(comp_sol_fn)
    assert not err, f'Complete solution build error: {err}'

    result, tests_passed = test_solution(sol_fn_new, comp_sol_fn_new, tests)
    os.remove(sol_fn_new)
    os.remove(comp_sol_fn_new)
    result['tests_passed'] = tests_passed
    return result


def run_test(solution):
    sol_fn = save_file(solution, 'user_solution')
    comp_sol_fn = save_file(solution, 'complete_solution')
    assert sol_fn and comp_sol_fn, 'Internal error: malformed runner data'

    tests = {}
    tests_total = 0 # If there are no tests, then use at least one (w/o params)
    tests = solution['tests']
    for t in tests.keys():
        tests[t] = tests[t].split('\n')
        tests_total += len(tests[t])
    if not tests_total:
        tests_total = 1

    result = run_tests(sol_fn, comp_sol_fn, tests)
    if 'error' in result:
        err = result.pop('error')
        tests_passed = result.pop('tests_passed')
        return {
            'error_data': {
                err: result,
                'tests_passed': tests_passed,
                'tests_total': tests_total,                
            }
        }

    assert tests and tests_total == result.get('tests_passed'), \
        f'Tests count does not match: total {tests_total}, passed {result.get("tests_passed")}'
    del result['tests_passed']
    return result

conn = redis.StrictRedis(
        host=os.getenv('REDIS_HOST'),
        port=os.getenv('REDIS_PORT'),
        db=0,
        username=os.getenv('REDIS_RUNNER_USER'),
        password=os.getenv('REDIS_RUNNER_PASSWORD'))
tasks = os.getenv('REDIS_TASK_LIST')
results = os.getenv('REDIS_RESULT_LIST')

task_schema = Task()
result_schema = Result()
while True:
    task_json = None
    try:
        _, task_json = conn.brpop(tasks)
        task = task_schema.loads(task_json)
        result = run_test(task)
        result['id'] = task['id']
        conn.lpush(results, result_schema.dumps(result))
    except Exception as e:
        logging.error(f'Exception for solution: {task_json if task_json else None} - {e}')
        conn.lpush(results, json.dumps({
            'id': task['id'],
            'internal_error': str(e)
        }))

