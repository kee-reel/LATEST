import sys
import time
import random
from db_helper import open_db

last_name, name, number, group = sys.argv[1:]

conn = open_db()
cur = conn.cursor()
cur.execute('select id from users where (num = %s AND group_name = %s)', (number, group))
rows = cur.fetchall()

if not rows:
    random.seed(time.time())
    data = (number, group, name, last_name)
    print(f'Adding user: {data}')
    cur.execute('insert into users(num, group_name, name, last_name) values(%s, %s, %s, %s)', data)
    conn.commit()

