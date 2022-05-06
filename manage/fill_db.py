import os
import re
import sys
import json
from glob import glob
from pathlib import Path
from collections import OrderedDict

from db_helper import open_db

LANGS = {}

def parse_filename(filename):
    return tuple(re.findall(r'(.*?)/', filename)[1:])


def upsert(cur, table, keys, data, return_id=True):
    all_data = keys.copy()
    all_data.update(data)
    query = f''' insert into {table}({ ','.join(all_data.keys()) })
            values({ ','.join(['%s']*len(all_data.values())) })
        on conflict({ ','.join(keys) }) do update 
            set { ','.join(f'{k}=%s' for k in data.keys()) }
        {'returning id' if return_id else ''}'''
    values = list(all_data.values())
    values.extend(data.values())
    cur.execute(query, values)
    return cur.fetchone()[0] if return_id else None


def add_project(cur, desc, path, folder):
    keys = {
        'folder_name': folder
    }
    data = {
        'name': desc['name']
    }
    return upsert(cur, 'projects', keys, data)


def add_unit(cur, desc, path, project_id, folder):
    keys = {
        'project_id': project_id,
        'folder_name': folder
    }
    data = {
        'name': desc['name']
    }
    return upsert(cur, 'units', keys, data)


def add_task(cur, desc, path, project_id, unit_id, folder):
    folder_path = path[:path.rindex('/')]
    files = os.listdir(folder_path)
    extention = None
    for f in files:
        if 'complete_solution' in f:
            extention = f[f.rindex('.')+1:]
            break
    assert extention, f'[ERROR] No complete_solution for {path}'

    keys = {
        'project_id': project_id,
        'unit_id': unit_id,
        'folder_name': folder
    }
    fixed_tests = open(f'{folder_path}/fixed_tests.txt', 'r').read()
    fixed_tests = '\n'.join(filter(lambda s: bool(s), fixed_tests.split('\n')))
    data = {
        'name': desc['name'],
        'position': desc.get('position', 0),
        'language_id': LANGS[extention],
        'description': desc['desc'],
        'input': json.dumps(desc['input'], ensure_ascii=False),
        'output': desc['output'],
        'source_code': open(f'{folder_path}/complete_solution.{extention}', 'r').read(),
        'fixed_tests': fixed_tests,
        'score': desc['score'],
    }
    id_ = upsert(cur, 'tasks', keys, data)

    for l in LANGS:
        fn = f'{folder_path}/template.{l}'
        if not os.path.exists(fn):
            continue
        upsert(cur, 'solution_templates', {
            'task_id': id_,
            'language_id': LANGS[l]
        }, {
            'source_code': open(fn, 'r').read()
        }, False)
    return id_


def add_template(cur, path):
    extention = path[path.rindex('.')+1:]

    keys = {
        'extention': extention,
    }
    data = {
        'template_solution': open(path, 'r').read(),
    }
    LANGS[extention] = upsert(cur, 'languages', keys, data)


conn = open_db()
cur = conn.cursor()

type_to_paths = OrderedDict()
expansion = '/*'
for t in ('project', 'unit', 'task'):
    paths = glob(f'tests{expansion}/desc.json')
    paths.sort()
    type_to_paths[t] = paths
    expansion += '/*'

for f in glob('tests/templates/*'):
    add_template(cur, f)

folders_to_id = {}
for t, paths in type_to_paths.items():
    for p in paths:
        print(f'Adding {p}')
        folders = parse_filename(p)

        folders_data = list(folders)
        for i in range(len(folders)-1, 0, -1):
            folders_data[i-1] = folders_to_id[folders[:i]]
            assert folders_data[i-1]

        data = json.loads(open(p, 'r', encoding='utf-8').read(), strict=False)
        id_ = globals().get(f'add_{t}')(cur, data, p, *folders_data)
        folders_to_id[folders] = id_

conn.commit()


