import os
import re
import sys
import json
from db_helper import open_db


def parse_filename(filename):
    return re.findall(r'(.*?)/', filename)[1:]


def fill_db(data, path, project, unit=None, task=None):
    conn = open_db()
    cur = conn.cursor()
    cur.execute('select id from projects where folder_name = %s', (project,))
    rows = cur.fetchone()
    project_id = rows[0] if rows else None
    name = data['name'] if unit is None else ''
    if project_id and unit is None:
        cur.execute('update projects set name = %s where id = %s', (name, project_id))
    elif not project_id:
        cur.execute('''insert into projects(folder_name, name) values(%s, %s) returning id''', (project, name))
        project_id = cur.fetchone()

    if unit is None:
        return
    cur.execute('select id from units where folder_name = %s AND project_id = %s', (unit, project_id))
    rows = cur.fetchone()
    unit_id = rows[0] if rows else None
    name = data['name'] if task is None else ''
    next_unit_id = data.get('next', None) if task is None else None
    if unit_id and task is None:
        cur.execute('update units set next_unit_id = %s, project_id = %s, name = %s, folder_name = %s where id = %s', 
                (next_unit_id, project_id, name, unit, unit_id))
    elif not unit_id:
        cur.execute('''insert into units(folder_name, next_unit_id, project_id, name)
            values(%s, %s, %s, %s) returning id''',
                (unit, next_unit_id, project_id, name))
        unit_id = cur.fetchone()

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

    cur.execute('select id from tasks where project_id = %s AND unit_id = %s AND folder_name = %s', (project_id, unit_id, task))
    rows = cur.fetchone()
    task_id = rows[0] if rows else None
    if task_id:
        cur.execute('''update tasks set project_id = %s, unit_id = %s, position = %s, folder_name = %s,
                extention = %s, name = %s, description = %s, input = %s, output = %s, source_code = %s, 
                fixed_tests = %s where id = %s''', (
            project_id,
            unit_id,
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
        cur.execute('''insert into tasks(project_id, unit_id, position, folder_name,
                    extention, name, description, input, output, source_code, fixed_tests) 
                values(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)''', (
            project_id,
            unit_id,
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
