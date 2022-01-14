import sys
import time
import random
import sqlite3 as db

last_name, name, number, group = sys.argv[1:]

conn = db.connect('tasks.db')
cur = conn.cursor()
cur.execute('select id from user where number = ? AND group_name = ?', (number, group))
rows = cur.fetchall()

if not rows:
    random.seed(time.time())
    data = (number, group, name, last_name)
    print(f'Adding user: {data}')
    cur.execute('insert into user(number, group_name, name, last_name) values(?, ?, ?, ?)', data)
    conn.commit()

