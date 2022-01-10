#!/bin/bash
curl http://localhost:1234?token=$(sqlite3 tasks.db "select token from access_token where user_id = 1")

