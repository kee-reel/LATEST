import re
import sys
import json
import sqlite3 as db


def parse_filename(filename):
    res = re.search(r'subject-(\d+)/work-(\d+)/variant-(\d+)/task-(\d+)', filename).groups()
    assert len(res) == 4, f'Filename "{filename}" have wrong pattern'
    return res


def process(filename):
    subject, work, variant, task = parse_filename(filename)
    with open(filename, 'r', encoding='utf-8') as f:
        data = f.read()
        data = json.loads(data)
        conn = db.connect('tasks.db')
        cur=conn.cursor()
        path = re.search(r'(.+task-\d+/)', filename).groups()[0]
        cur.execute('select id from task where subject = ? AND work = ? AND variant = ? AND number = ?', (subject, work, variant, task))
        rows = cur.fetchall()
        if rows:
            task_id = rows[0][0]
            cur.execute('update task set subject = ?, work = ?, variant = ?, number = ?, name = ?, desc = ?, input = ?, output = ? where id = ?', (
                subject,
                work,
                variant,
                task,
                data['name'],
                data['desc'],
                json.dumps(data['input'], ensure_ascii=False),
                data['output'],
                task_id)
            )
        else:
            cur.execute('insert into task(subject, work, variant, number, name, desc, input, output) values(?, ?, ?, ?, ?, ?, ?, ?)', (
                subject,
                work,
                variant,
                task,
                data['name'],
                data['desc'],
                json.dumps(data['input'], ensure_ascii=False),
                data['output'])
            )
        conn.commit()


if len(sys.argv) >= 2:
    filenames = [sys.argv[1]]
else:
    filenames = sys.stdin.readlines()

for f in filenames:
    process(f.replace('\n', ''))
