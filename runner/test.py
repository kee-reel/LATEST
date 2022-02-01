#!/usr/bin/env python3
import os
import sys
import subprocess


LANG_TO_EXEC = {
    'exe': lambda filename: [f'./{filename}'],
    'py': lambda filename: ['python3', f'{filename}']
}


def prepare_str(s):
    if s[-1] == '\n':
        s = s[:-1]
    try:
        last_line_index = s.rindex('\n') + 1
    except ValueError as e:
        last_line_index = 0
    return s[last_line_index:]


def execute(cmd, params):
    try:
        p = subprocess.run(cmd, input=params, capture_output=True, text=True, timeout=3, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, e.stderr


def run(exec_cmd, sol, comp_sol, is_verbose, params=None):
    exit_code = 0
    error = None

    expected, err = execute(exec_cmd(comp_sol), params)
    if err:
        return None, {
            'error': err,
            'result': expected
        }

    result, err = execute(exec_cmd(sol), params)
    if err:
        return None, {
            'error': err,
            'result': result
        }

    expected = prepare_str(expected)
    result = prepare_str(result)
    if expected != result:
        return None, {
            'expected': expected,
            'result': result
        }
    elif is_verbose:
        return {
            'result': result
        }, None
    return None, None

def test_solution(solution, complete_solution, test_sets, is_verbose):
    extention = complete_solution[complete_solution.rindex('.')+1:]
    assert extention in LANG_TO_EXEC, 'Language not supported'
    exec_cmd = LANG_TO_EXEC[extention]

    if not os.path.exists(solution) or not os.path.exists(complete_solution):
        return {
            'err': f'Required files not found'
        }

    is_tested = False
    results = []
    error = None
    for tests in test_sets.values():
        for test in tests.split('\n'):
            if not test:
                continue
            cmd_line = test.replace(';', '\n')
            result, error = run(exec_cmd, solution, complete_solution, is_verbose, cmd_line)

            is_tested = True
            if result:
                result['params'] = test
                results.append(result)
            if error:
                error['params'] = test
                break
        if error:
            break

    # No test cases
    if not is_tested and not error:
        result, error = run(exec_cmd, solution, complete_solution, is_verbose)
        if result:
            results.append(result)

    return results, error

