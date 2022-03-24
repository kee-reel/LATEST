import os
import sys
import subprocess


LANGS = ['c', 'py']
LANG_CMD = {
        'c': lambda source, target: ['/usr/bin/gcc', source, '-o', target, '-lm'],
}


def execute(cmd):
    try:
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=3, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, {'msg': e.stderr}


def build_solution(solution):
    extention = solution[solution.rindex('.')+1:]
    assert extention in LANGS, 'Language not supported'
    if extention not in LANG_CMD:
        return solution, None

    compiled_solution = solution[:solution.rindex('.')] + '.exe'
    cmd = LANG_CMD[extention](solution, compiled_solution)
    _, err = execute(cmd)
    return compiled_solution, err

