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
from schemas import TestResult, Solution

fn_letters = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'

def save_file(solution, field):
    data = solution[field]

    text = data['text']
    extention = data['extention']
    if not extention or not text:
        return None

    random.seed(time.time())
    postfix = ''.join(random.choice(fn_letters) for _ in range(8))
    fn = os.path.join(PATH, f'{field}_{postfix}.{extention}')

    if text:
        open(fn, 'w').write(text)
    else:
        file_.save(fn)

    return fn


def run_tests(sol_fn, comp_sol_fn, tests, is_verbose):
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

    result, tests_passed = test_solution(sol_fn_new, comp_sol_fn_new, tests, is_verbose)
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

    result = run_tests(sol_fn, comp_sol_fn, tests, solution['verbose'])
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

conn = redis.Redis(os.getenv('REDIS_HOST'), os.getenv('REDIS_PORT'))
solutions = os.getenv('REDIS_SOLUTIONS_LIST_PREFIX')
tests = os.getenv('REDIS_TESTS_LIST')

solution_schema = Solution()
test_result_schema = TestResult()
while True:
    try:
        _, solution_json = conn.brpop(solutions)
        solution = json.loads(solution_json)
        err = solution_schema.validate(solution)
        if err:
            raise Exception(err)
        test_result = run_test(solution)
        test_result['id'] = solution['id']
        err = test_result_schema.validate(test_result)
        if err:
            raise Exception(err)
        conn.lpush(tests, json.dumps(test_result))
    except Exception as e:
        logging.error(f'Exception for solution: {solution_json} - {e}')
        conn.lpush(tests, json.dumps({
            'id': solution['id'],
            'internal_error': str(e)
        }))

