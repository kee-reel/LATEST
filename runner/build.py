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
TRIGGERS = {
        'c': ('system',),
        'cpp': ('system',),
        'pas': ('system',),
        'py': ('import os', 'import subprocess', 'exec\\('),
}
for t, l in TRIGGERS.items():
    TRIGGERS[t] = '|'.join(l)
TIMEOUT = 5


def execute(cmd):
    try:
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=TIMEOUT, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, e.stderr


def build_solution(solution):
    ext = solution[solution.rindex('.')+1:]
    assert ext in LANGS, 'Language not supported'
    if ext in LANG_CMD:
        compiled_solution = solution[:solution.rindex('.')] + '.exe'
        cmd = LANG_CMD[ext](solution, compiled_solution)
        _, err = execute(cmd)
        if err:
            start_word = None
            if ext == 'pas':
                start_word = 'Fatal: '
            elif ext == 'py':
                start_word = 'line '
            elif ext == 'c' or ext == 'cpp':
                start_word = 'error'
            if start_word:
                i = err.find(start_word)
                if i != -1:
                    err = err[i:]
            return None, {
                "error": ERROR.BUILD,
                "msg": err,
            }
    else:
        compiled_solution = solution

    out, err = execute(['grep', '-oaE', TRIGGERS[ext], compiled_solution])
    if out or err:
        return None, {
            "error": ERROR.BUILD,
            "msg": 'System calls are no allowed'
        }
    return compiled_solution, err

