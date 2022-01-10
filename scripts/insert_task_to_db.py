import re
import sys
import json
import sqlite3 as db


def parse_filename(filename):
    return re.search(r'(subject-(\d+)/)(work-(\d+)/)?(variant-(\d+)/)?(task-(\d+))?', filename).groups()


def fill_db(cur, data, subject, work, variant, task):
    cur.execute('select id from subject where id = ?', (subject,))
    rows = cur.fetchone()
    name = data['name'] if work is None else ''
    if rows and work is None:
        cur.execute('update subject set name = ? where id = ?', (name, subject))
    elif not rows:
        cur.execute('insert into subject(id, name) values(?, ?)', (subject, name))

    if work is None:
        return
    cur.execute('select id from work where id = ?', (work,))
    rows = cur.fetchone()
    name = data['name'] if variant is None else ''
    next_work_id = data.get('next', None) if variant is None else None
    if rows and variant is None:
        cur.execute('update work set next_work_id = ?, subject = ?, name = ? where id = ?', 
                (next_work_id, subject, name, work))
    elif not rows:
        cur.execute('insert into work(id, next_work_id, subject, name) values(?, ?, ?, ?)',
                (work, next_work_id, subject, name))

    if variant is None:
        return
    cur.execute('select id from variant where id = ?', (variant,))
    rows = cur.fetchone()
    name = data['name'] if task is None else ''
    if rows and task is None:
        cur.execute('update variant set work = ?, subject = ?, name = ? where id = ?', (work, subject, name, variant))
    elif not rows:
        cur.execute('insert into variant(id, work, subject, name) values(?, ?, ?, ?)', 
                (variant, work, subject, name))

    if task is None:
        return
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


def process(filename):
    filename_data = parse_filename(filename)
    subject, work, variant, task = filename_data[1], filename_data[3], filename_data[5], filename_data[7]
    if subject is None:
        return

    data = None
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.loads(f.read())

    conn = db.connect('tasks.db')
    cur=conn.cursor()
    fill_db(cur, data, subject, work, variant, task)
    conn.commit()


if len(sys.argv) >= 2:
    filenames = [sys.argv[1]]
else:
    filenames = sys.stdin.readlines()

for f in filenames:
    process(f.replace('\n', ''))
