from db_helper import open_db

conn = open_db()
cur = conn.cursor()
cur.execute('truncate table tasks')
cur.execute('truncate table works')
cur.execute('truncate table subjects')
conn.commit()
