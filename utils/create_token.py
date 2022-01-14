#!/usr/bin/python3
import sys
import time
import random
import sqlite3 as db


def main():
    data = sys.argv[1:]
    if len(data) == 2:
        subject = data[0]
        user = variant = None
    elif len(data) == 3:
        subject, variant, user = data
        variant = int(variant)
        user = int(user)
    else:
        print(f'Usage: {sys.argv[0]} SUBJECT_ID [VARIANT_ID USER_ID]')
        return

    conn = db.connect('tasks.db')
    cur = conn.cursor()

    if user is None:
        cur.execute('select id, number from user')
        users = cur.fetchall()
        if not users:
            print('No users found')
            return
    else:
        users = [(user, variant)]
    print(users)

    cur.execute('select t.variant from task as t where t.subject = ? GROUP BY t.variant', (subject,))
    rows = cur.fetchall()
    if not rows:
        print('Unknown work')
        return
    variant_ids = [row[0] for row in rows]
    print(variant_ids)
    
    s = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    for user_id, number in users:
        random.seed(time.time() * 3)
        token = ''.join(random.choice(s) for _ in range(256))
        n = len(variant_ids)
        variant = number % n if number > n else number
        data = (token, user_id, subject, variant)
        print(f'Adding user: {data}')
        cur.execute('insert into access_token(token, user_id, subject, variant) values(?, ?, ?, ?)', data)
        time.sleep(random.random())
    conn.commit()

main()
