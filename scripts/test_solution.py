#!/usr/bin/env python3
import os
import sys
import json
import subprocess


SOLUTION = './solution'
COMPLETE_SOLUTION = './complete_solution'


def run(program_name, params=None):
    exit_code = 0
    error = None
    data = {'params': params if params else ''}
    try:
        p = subprocess.run([COMPLETE_SOLUTION], input=params, capture_output=True, text=True, timeout=3, user='testrun', check=True)
        expected = p.stdout
        expected = expected[expected.rindex('\n')+1:]

        p = subprocess.run([program_name], input=params, capture_output=True, text=True, timeout=3, user='testrun', check=True)
        result = p.stdout
        result = result[result.rindex('\n')+1:]
        data['result'] = result

        if expected != result:
            data['expected'] = expected
            error = 'not_equal'
    except subprocess.CalledProcessError as e:
        error = e.stderr
        result['result'] = e.stdout
    return result, error

def main():
    path = sys.argv[1]
    os.chdir(path)
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
                data, error = run(SOLUTION, cmd_line)
                is_tested = True
                data = {
                    'in': line,
                    'out': data
                }
                if error:
                    result = [data]
                    break
                result.append(data)
                line = f.readline()

    # No test cases
    if not is_tested and not error:
        data, error = run(SOLUTION)
        result.append({
            'in': '',
            'out': data
        })

    print(json.dumps({'results': result, 'error': error}))

main()
