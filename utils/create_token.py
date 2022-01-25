#!/usr/bin/python3
import sys
import time
import random
from db_helper import open_db


def main():
    data = sys.argv[1:]
    if len(data) == 2:
        email, project = data
    else:
        print(f'Usage: {sys.argv[0]} EMAIL PROJECT_ID')
        return

    conn = open_db()
    cur = conn.cursor()

    cur.execute('select id from users where email = %s', (email,))
    row = cur.fetchone()
    if not row:
        print('No users found')
        return
    user_id = row[0]

    random.seed(time.time())
    s = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    s = ''.join(random.sample(s,len(s)))
    token = ''.join(random.choice(s) for _ in range(256))
    data = (token, user_id, project)

    cur.execute('insert into tokens(token, user_id, project_id) values(%s, %s, %s)', data)
    conn.commit()

main()
