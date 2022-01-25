import os
import psycopg2
from dotenv import load_dotenv


def env(name):
    return os.environ[name]


def open_db():
    return psycopg2.connect(database = env('DB_NAME'), user = env('DB_USER'), 
        password = env('DB_PASS'), host = env('DB_HOST'), port = env('DB_PORT'))

load_dotenv()

