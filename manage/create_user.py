import sys
import time
import random
from db_helper import open_db

nick, email = sys.argv[1:]

conn = open_db()
cur = conn.cursor()
cur.execute('select id from users where (email = %s)', (email,))
rows = cur.fetchall()

if not rows:
    random.seed(time.time())
    data = (email, nick)
    print(f'Adding user: {data}')
    cur.execute('insert into users(email, nick) values(%s, %s)', data)
    conn.commit()

