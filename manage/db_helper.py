import os
import psycopg2


def env(name):
    return os.environ[name]


def open_db():
    return psycopg2.connect(
        database = env('POSTGRES_DB'),
        user = env('POSTGRES_USER'), 
        password = env('POSTGRES_PASSWORD'),
        host = env('DB_HOST'),
        port = env('DB_PORT')
    )

