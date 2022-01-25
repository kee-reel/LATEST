import os
import re
import sys
import json
from db_helper import open_db


def parse_filename(filename):
    return re.findall(r'(.*?)/', filename)[1:]


def fill_db(data, path, subject, work=None, task=None):
    conn = open_db()
    cur = conn.cursor()
    cur.execute('select id from subjects where folder_name = %s', (subject,))
    rows = cur.fetchone()
    subject_id = rows[0] if rows else None
    name = data['name'] if work is None else ''
    if subject_id and work is None:
        cur.execute('update subjects set name = %s where id = %s', (name, subject_id))
    elif not subject_id:
        cur.execute('''insert into subjects(folder_name, name) values(%s, %s) returning id''', (subject, name))
        subject_id = cur.fetchone()

    if work is None:
        return
    cur.execute('select id from works where folder_name = %s AND subject_id = %s', (work, subject_id))
    rows = cur.fetchone()
    work_id = rows[0] if rows else None
    name = data['name'] if task is None else ''
    next_work_id = data.get('next', None) if task is None else None
    if work_id and task is None:
        cur.execute('update works set next_work_id = %s, subject_id = %s, name = %s, folder_name = %s where id = %s', 
                (next_work_id, subject_id, name, work, work_id))
    elif not work_id:
        cur.execute('''insert into works(folder_name, next_work_id, subject_id, name)
            values(%s, %s, %s, %s) returning id''',
                (work, next_work_id, subject_id, name))
        work_id = cur.fetchone()

    if task is None:
        return

    folder_path = path[:path.rindex('/')]
    files = os.listdir(folder_path)
    extention = None
    for f in files:
        if 'complete_solution' in f:
            extention = f[f.rindex('.')+1:]
            break
    if not extention:
        print(f'[ERROR] No complete_solution for {path}', file=sys.stderr)
        return

    fixed_tests_path = f'{folder_path}/fixed_tests.txt'
    source_code_path = f'{folder_path}/complete_solution.{extention}'
    if not os.path.exists(fixed_tests_path) or not os.path.exists(source_code_path):
        print(f'[ERROR] No fixed tests or complete solution provided for {path}')
        return

    task_pos = data.get('position', 0)
    fixed_tests = open(fixed_tests_path, 'r').read()
    source_code = open(source_code_path, 'r').read()

    cur.execute('select id from tasks where subject_id = %s AND work_id = %s AND folder_name = %s', (subject_id, work_id, task))
    rows = cur.fetchone()
    task_id = rows[0] if rows else None
    if task_id:
        cur.execute('''update tasks set subject_id = %s, work_id = %s, position = %s, folder_name = %s,
                extention = %s, name = %s, description = %s, input = %s, output = %s, source_code = %s, 
                fixed_tests = %s where id = %s''', (
            subject_id,
            work_id,
            task_pos,
            task,
            extention,
            data['name'],
            data['desc'],
            json.dumps(data['input'], ensure_ascii=False),
            data['output'],
            source_code,
            fixed_tests,
            task_id)
        )
    else:
        print(f'With work {work_id}')
        cur.execute('''insert into tasks(subject_id, work_id, position, folder_name,
                    extention, name, description, input, output, source_code, fixed_tests) 
                values(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)''', (
            subject_id,
            work_id,
            task_pos,
            task,
            extention,
            data['name'],
            data['desc'],
            json.dumps(data['input'], ensure_ascii=False),
            data['output'],
            source_code,
            fixed_tests)
        )
    conn.commit()


def process(filename):
    filename_data = parse_filename(filename)
    if not filename_data:
        print(f'Malformed path: {filename}', file=sys.stderr)
        return

    data = None
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.loads(f.read())

    fill_db(data, filename, *filename_data)


if len(sys.argv) >= 2:
    filenames = [sys.argv[1]]
else:
    filenames = sys.stdin.readlines()

for f in filenames:
    process(f.replace('\n', ''))
