#!/usr/bin/python3
import sys
import time
import random
from db_helper import open_db


def main():
    data = sys.argv[1:]
    if len(data) == 1:
        subject = data[0]
        user = None
    elif len(data) == 2:
        subject, user = data
        user = int(user)
    else:
        print(f'Usage: {sys.argv[0]} SUBJECT_ID [VARIANT_ID USER_ID]')
        return

    conn = open_db()
    cur = conn.cursor()

    if user is None:
        cur.execute('select id, number from users')
        users = cur.fetchall()
        if not users:
            print('No users found')
            return
    else:
        users = [user]
    print(users)

    random.seed(time.time())
    s = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    s = ''.join(random.sample(s,len(s)))
    for user_id in users:
        random.seed(time.time())
        token = ''.join(random.choice(s) for _ in range(256))
        data = (token, user_id, subject)
        print(f'Adding user: {data}')
        cur.execute('insert into tokens(token, user_id, subject_id) values(%s, %s, %s)', data)
        time.sleep(random.random())
    conn.commit()

main()
