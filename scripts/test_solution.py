#!/usr/bin/env python3
import os
import sys
import json
import subprocess

SOLUTION = None
COMPLETE_SOLUTION = None


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

        if expected != result:
            return {
                'error': 'not_equal',
                'expected': expected,
                'result': result
            }
    except subprocess.CalledProcessError as e:
        return {
            'error': e.stderr,
            'result': e.stdout
        }
    return None

def main():
    path = sys.argv[1]
    global SOLUTION
    global COMPLETE_SOLUTION
    SOLUTION = sys.argv[2]
    COMPLETE_SOLUTION = sys.argv[2]
    extention = COMPLETE_SOLUTION[COMPLETE_SOLUTION.rindex('.')+1:]
    assert extention in LANG_TO_EXEC, 'Unsuppoted extention'
    exec_cmd = LANG_TO_EXEC[extention]
    os.chdir(path)
    files = os.listdir('.')
    assert SOLUTION in files and COMPLETE_SOLUTION in files, \
        'Target directory doesn\'t have required files'
    if extention == 'exe':
        SOLUTION = './' + SOLUTION
        COMPLETE_SOLUTION = './' + COMPLETE_SOLUTION
    test_types = ['user', 'fixed', 'random']
    is_tested = False
    data = {}
    error = None
    for t in test_types:
        if error:
            break
        with open(f'{t}_tests.txt', 'r') as f:
            line = f.readline()
            while line != '':
                cmd_line = line.replace(';', '\n')
                error = run(exec_cmd, cmd_line)
                is_tested = True
                if error:
                    error['params'] = line[:-1] if line[-1] == '\n' else line
                    break
                line = f.readline()

    # No test cases
    if not is_tested and not error:
        error = run(exec_cmd)

    if error:
        data['error'] = error

    print(json.dumps(data))

main()
