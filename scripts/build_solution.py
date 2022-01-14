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
        return e.stdout, e.stderr


def main():
    path = sys.argv[1]
    solution = sys.argv[2]
    complete_solution = sys.argv[3]

    extention = complete_solution[complete_solution.rindex('.')+1:]
    if extention not in COMPILED_LANGS:
        print(extention, end='')
        return

    assert extention in SUPPORTED_COMPILED_LANGS, 'Language not supported'
    solution_without_ext = solution[:solution.rindex('.')]
    complete_solution_without_ext = complete_solution[:complete_solution.rindex('.')]

    os.chdir(path)
    _, err = execute(['/usr/bin/gcc', '-o', f'{complete_solution_without_ext}.exe', complete_solution, '-lm'])
    if err:
        print(err, file=sys.stderr)
    _, err = execute(['/usr/bin/gcc', '-o', f'{solution_without_ext}.exe', solution, '-lm'])
    if err:
        print(err, file=sys.stderr)
    print('exe', end='')

main()

