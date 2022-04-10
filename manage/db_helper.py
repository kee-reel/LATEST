import os
import psycopg2


def env(name):
    return os.environ[name]


def open_db():
    return psycopg2.connect(database = env('DB_NAME'), user = env('DB_USER'), 
        password = env('DB_PASS'), host = env('DB_HOST'), port = 5432)

