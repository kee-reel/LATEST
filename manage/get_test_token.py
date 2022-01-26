import sys
from db_helper import open_db

conn = open_db()
cur = conn.cursor()
cur.execute('select token from tokens where user_id = 1')
data = cur.fetchone()
token = data[0] if data else None
assert token, 'Database have no token'
print(token, end='')

