#!/usr/env python3
import os
import sys
import time
import random
import logging
from build import build_solution
from test import test_solution
from flask import Flask, request, url_for

app = Flask(__name__)
app.logger.setLevel(logging.INFO)

fn_letters = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
UPLOAD_FOLDER = './uploads'


def save_file(request, field):
    extention = request.form.get('extention')

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


@app.route('/', methods=['POST'])
def run_test():
    is_verbose = request.form.get('verbose') == 'true'
    sol_fn = save_file(request, 'solution')
    comp_sol_fn = save_file(request, 'complete_solution')
    if not sol_fn or not comp_sol_fn:
        return {'error': 'Internal error: malformed runner data'}

    tests = {}
    for t in ('user', 'fixed', 'random'):
        test_set = request.form.get(f'{t}_tests')
        if test_set:
            tests[t] = test_set

    sol_fn_new, comp_sol_fn_new, err = build_solution(sol_fn, comp_sol_fn)
    if sol_fn != sol_fn_new:
        os.remove(sol_fn)
        os.remove(comp_sol_fn)
    if err:
        return {'error': err}

    results, err = test_solution(sol_fn_new, comp_sol_fn_new, tests, is_verbose)

    os.remove(sol_fn_new)
    os.remove(comp_sol_fn_new)
    data = {'error': err}
    if results:
        data['results'] = results
    return data

app.run(host='0.0.0.0', port=1337)

