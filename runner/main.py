#!/usr/env python3
import os
import sys
import time
import random
import logging
from build import build_solution, LANGS
from test import test_solution
from flask import Flask, request, url_for

app = Flask(__name__)
app.logger.setLevel(logging.DEBUG)

fn_letters = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
UPLOAD_FOLDER = './uploads'

def save_file(request, field):
    extention = request.form.get(f'{field}_ext')

    file_ = request.files.get(field)
    text = request.form.get(field)
    if not extention or (not file_ and not text):
        return None

    random.seed(time.time())
    postfix = ''.join(random.choice(fn_letters) for _ in range(8))
    fn = os.path.join(UPLOAD_FOLDER, f'{field}_{postfix}.{extention}')

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


@app.route('/', methods=['GET', 'POST'])
def run_test():
    if request.method == 'GET':
        return {'langs': LANGS}
    is_verbose = request.form.get('verbose') == 'true'
    sol_fn = save_file(request, 'solution')
    comp_sol_fn = save_file(request, 'complete_solution')
    assert sol_fn and comp_sol_fn, 'Internal error: malformed runner data'

    tests = {}
    tests_total = 0
    for t in ('user', 'fixed', 'random'):
        test_set = request.form.get(f'{t}_tests')
        if test_set:
            test_set = test_set.split('\n')
            tests[t] = test_set
            tests_total += len(test_set)
    if not tests:
        tests_total = 1

    result = run_tests(sol_fn, comp_sol_fn, tests, is_verbose)
    result['tests_total'] = tests_total
    if 'error' in result:
        return {'error_data': result}
    elif tests:
        assert tests_total == result.get('tests_passed'), \
            f'Tests count does not match: total {tests_total}, passed {result.get("tests_passed")}'
    return result

app.run(host='0.0.0.0', port=os.getenv('RUNNER_PORT'))

