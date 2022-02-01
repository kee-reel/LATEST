import os
import sys
import subprocess


COMPILED_LANGS = ['c', 'cpp', 'go', 'cs']
SUPPORTED_COMPILED_LANGS = ['c']


def execute(cmd):
    try:
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=3, check=True)
        return p.stdout, None
    except subprocess.CalledProcessError as e:
        return e.stdout, {'msg': e.stderr}


def build_solution(solution, complete_solution):
    extention = complete_solution[complete_solution.rindex('.')+1:]
    if extention not in COMPILED_LANGS:
        return solution, complete_solution, None

    assert extention in SUPPORTED_COMPILED_LANGS, 'Language not supported'
    sol_wo_ext = solution[:solution.rindex('.')]
    comp_sol_wo_ext = complete_solution[:complete_solution.rindex('.')]

    _, err = execute(['/usr/bin/gcc', '-o', f'{comp_sol_wo_ext}.exe', complete_solution, '-lm'])
    if err:
        return None, None, err
    _, err = execute(['/usr/bin/gcc', '-o', f'{sol_wo_ext}.exe', solution, '-lm'])
    if err:
        return None, None, err
    return f'{sol_wo_ext}.exe', f'{comp_sol_wo_ext}.exe', None

