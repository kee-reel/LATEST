from db_helper import open_db

conn = open_db()
cur = conn.cursor()
cur.execute('truncate table tasks')
cur.execute('truncate table units')
cur.execute('truncate table projects')
conn.commit()
