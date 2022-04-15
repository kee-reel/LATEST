#!/usr/bin/env python3
import os
import sys
import subprocess
from errors import ERROR

RUN_TIMEOUT = 0.5
LANG_TO_EXEC = {
    'exe': lambda filename: [f'./{filename}'],
    'py': lambda filename: ['python3', f'{filename}']
}


def get_ext(f):
    return f[f.rindex('.')+1:]


def prepare_str(s):
    if not s:
        return s
    if s[-1] == '\n':
        s = s[:-1]
    try:
        last_line_index = s.rindex('\n') + 1
    except ValueError as e:
        last_line_index = 0
    return s[last_line_index:]


def execute(cmd, params):
    try:
        p = subprocess.run(cmd, input=params, capture_output=True, text=True, timeout=RUN_TIMEOUT, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, e.stderr


def run(sol, comp_sol, params=None):
    try:
        result, err = execute(sol, params)
    except subprocess.TimeoutExpired:
        return {
            'error': ERROR.TIMEOUT,
        }
    if err:
        return {
            'error': ERROR.RUNTIME,
            'msg': err
        }

    try:
        expected, err = execute(comp_sol, params)
    except subprocess.TimeoutExpired:
        assert False, 'Timeout on execution of complete solution'
    assert not err, f'Error on execution of complete solution: {err}'

    expected = prepare_str(expected)
    result = prepare_str(result)
    if expected != result:
        return {
            'error': ERROR.TEST,
            'expected': expected,
            'result': result
        }
    return {'result': result}

def test_solution(solution, complete_solution, test_sets, is_verbose):
    sol_ext = get_ext(solution)
    comp_sol_ext = get_ext(complete_solution)
    assert sol_ext in LANG_TO_EXEC and comp_sol_ext in LANG_TO_EXEC, 'Language not supported'
    assert os.path.exists(solution) and os.path.exists(complete_solution), 'Not found solution files'

    is_tested = False
    results = []
    sol_exec = LANG_TO_EXEC[sol_ext](solution)
    comp_sol_exec = LANG_TO_EXEC[comp_sol_ext](complete_solution)
    tests_count = 0
    for type_, tests in test_sets.items():
        for test in tests:
            assert test, f'Empty test in {type_} tests: {tests}'
            cmd_line = test.replace(';', '\n')
            result = run(sol_exec, comp_sol_exec, cmd_line)

            is_tested = True
            result['params'] = test
            if 'error' in result:
                return result, tests_count
            if is_verbose:
                results.append(result)
            tests_count += 1

    # No test cases
    if not is_tested:
        result = run(sol_exec, comp_sol_exec, is_verbose)
        if 'error' in result:
            return result, tests_count
        if is_verbose:
            results.append(result)
        tests_count += 1

    return {'result': results} if results else {}, tests_count

