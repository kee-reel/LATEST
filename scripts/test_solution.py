#!/usr/bin/env python3
import os
import sys
import json
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


def run(exec_cmd, sol, comp_sol, params=None):
    exit_code = 0
    error = None

    expected, err = execute(exec_cmd(comp_sol), params)
    if err:
        return {
            'error': err,
            'result': expected
        }

    result, err = execute(exec_cmd(sol), params)
    if err:
        return {
            'error': err,
            'result': result
        }

    expected = prepare_str(expected)
    result = prepare_str(result)
    if expected != result:
        return {
            'error': 'not_equal',
            'expected': expected,
            'result': result
        }
    return None

def main():
    path = sys.argv[1]
    solution = sys.argv[2]
    complete_solution = sys.argv[3]

    extention = complete_solution[complete_solution.rindex('.')+1:]
    assert extention in LANG_TO_EXEC, 'Language not supported'
    exec_cmd = LANG_TO_EXEC[extention]

    os.chdir(path)
    files = os.listdir('.')
    assert solution in files and complete_solution in files, \
        'Target directory doesn\'t have required files'

    is_tested = False
    data = {}
    error = None
    for t in ('user', 'fixed', 'random'):
        with open(f'{t}_tests.txt', 'r') as f:
            line = f.readline()
            while line != '':
                cmd_line = line.replace(';', '\n')
                error = run(exec_cmd, solution, complete_solution, cmd_line)
                is_tested = True
                if error:
                    error['params'] = line[:-1] if line[-1] == '\n' else line
                    break
                line = f.readline()
        if error:
            break

    # No test cases
    if not is_tested and not error:
        error = run(sexec_cmd, solution, complete_solution)

    if error:
        data['error'] = error

    print(json.dumps(data))


main()

