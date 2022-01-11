#!/usr/bin/env python3
import os
import sys
import json
import subprocess

SOLUTION = 'solution.'
COMPLETE_SOLUTION = 'complete_solution.'


LANG_TO_EXEC = {
    'exe': None,
    'py': 'python3'
}

def run(exec_cmd, params=None):
    exit_code = 0
    error = None
    data = {}
    try:
        cmd = [COMPLETE_SOLUTION] if exec_cmd is None else [exec_cmd, COMPLETE_SOLUTION]
        p = subprocess.run(cmd, input=params, capture_output=True, text=True, timeout=3, check=True)
        expected = p.stdout
        if expected[-1] == '\n':
            expected = expected[:-1]
        try:
            last_line_index = expected.rindex('\n') + 1
        except ValueError as e:
            last_line_index = 0
        expected = expected[last_line_index:]

        cmd = [SOLUTION] if exec_cmd is None else [exec_cmd, SOLUTION]
        p = subprocess.run(cmd, input=params, capture_output=True, text=True, timeout=3, check=True)
        result = p.stdout
        if result[-1] == '\n':
            result = result[:-1]
        try:
            last_line_index = result[:-1].rindex('\n') + 1
        except ValueError as e:
            last_line_index = 0
        result = result[last_line_index:]
        data['result'] = result

        if expected != result:
            data['expected'] = expected
            error = 'not_equal'
    except subprocess.CalledProcessError as e:
        error = e.stderr
        data['result'] = e.stdout
    return data, error

def main():
    path = sys.argv[1]
    extention = sys.argv[2]
    assert extention in LANG_TO_EXEC, 'Unsuppoted extention'
    exec_cmd = LANG_TO_EXEC[extention]
    global SOLUTION
    global COMPLETE_SOLUTION
    SOLUTION += extention
    COMPLETE_SOLUTION += extention
    os.chdir(path)
    files = os.listdir('.')
    assert SOLUTION in files and COMPLETE_SOLUTION in files, \
        'Target directory doesn\'t have required files'
    if extention == 'exe':
        SOLUTION = './' + SOLUTION
        COMPLETE_SOLUTION = './' + COMPLETE_SOLUTION
    test_types = ['user', 'fixed', 'random']
    is_tested = False
    error = None
    result = []
    for t in test_types:
        if error:
            break
        with open(f'{t}_tests.txt', 'r') as f:
            line = f.readline()
            while line != '':
                cmd_line = line.replace(';', '\n')
                data, error = run(exec_cmd, cmd_line)
                data['params'] = line[:-1] if line[-1] == '\n' else line
                is_tested = True
                if error:
                    result = [data]
                    break
                result.append(data)
                line = f.readline()

    # No test cases
    if not is_tested and not error:
        data, error = run(exec_cmd)
        result.append(data)

    print(json.dumps({'results': result, 'error': error}))

main()
