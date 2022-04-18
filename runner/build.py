import os
import sys
import subprocess
from errors import ERROR


LANGS = ['c', 'cpp', 'py', 'pas']
LANG_CMD = {
        'c': lambda source, target: ['/usr/bin/gcc', source, '-o', target, '-lm'],
        'cpp': lambda source, target: ['/usr/bin/g++', source, '-o', target, '-lm'],
        'pas': lambda source, target: ['/usr/bin/fpc', '-ve', '-Fe/dev/stderr', source, f'-o{target}'],
}
TRIGGER_WORD = 'system'
TIMEOUT = 5


def execute(cmd):
    try:
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=TIMEOUT, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, e.stderr


def build_solution(solution):
    extention = solution[solution.rindex('.')+1:]
    assert extention in LANGS, 'Language not supported'
    if extention not in LANG_CMD:
        return solution, None

    compiled_solution = solution[:solution.rindex('.')] + '.exe'
    cmd = LANG_CMD[extention](solution, compiled_solution)
    _, err = execute(cmd)
    if err:
        return None, {
            "error": ERROR.BUILD,
            "msg": err,
        }
    out, err = execute(['grep', '-oe', TRIGGER_WORD, compiled_solution])
    if out or err:
        return None, {
            "error": ERROR.BUILD,
            "msg": 'System calls are no allowed',
        }
    return compiled_solution, err

