import os
import sys
import time
import random
import sqlite3 as db

subject, work = sys.argv[1:]

def main():
    conn = db.connect('tasks.db')
    cur = conn.cursor()

    cur.execute('''select a.token, u.group_name, u.number from task as t
            join user as u on u.id = a.user_id
            join access_token as a on a.variant = t.variant
            where t.subject = ? AND t.work = ? group by a.token''', (subject, work))
    tokens = cur.fetchall()
    if not tokens:
        print('No tokens found')
        return

    server = input('Mail server and port (HOST:PORT): ')
    login = input('Mail login: ')
    passwd = input('Mail passwd: ')

    cur.execute('select t.variant from task as t where t.subject = ? AND t.work = ? GROUP BY t.variant', (subject, work))
    rows = cur.fetchall()
    if not rows:
        print('Unknown work')
        return

    variant_ids = [row[0] for row in rows]
    n = len(variant_ids)

    receiver_postfix = input('Receivers mail postfix: ')

    for token, group_name, number in tokens:
        variant = number % n if number > n else number
        receiver = f'{group_name}b{number if number > 9 else str(0)+str(number)}@{receiver_postfix}'
        cmd = f'''sendemail -m "Varaint: {variant}
Token:
{token}" -f "{login}" -u "Autotests" -t "{receiver}" -s "{server}" -o tls=yes -xu "{login}"  -xp "{passwd}"'''
        print(cmd)
        os.system(cmd)

    return
    s = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    for user_id, number in users:
        random.seed(time.time() * 3)
        token = ''.join(random.choice(s) for _ in range(256))
        n = len(variant_ids)
        data = (token, user_id, subject, work, variant)
        print(f'Adding user: {data}')
        cur.execute('insert into access_token(token, user_id, subject, work, variant) values(?, ?, ?, ?, ?)', data)
        time.sleep(random.random())
    conn.commit()

main()
