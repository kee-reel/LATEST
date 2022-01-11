#!/bin/bash
TASK_ID=$1
TASK_FILE=$2
curl -F token=$(sqlite3 tasks.db "select token from access_token where user_id = 1") -F source_${TASK_ID}=@$(pwd)/$TASK_FILE -F tasks=$TASK_ID http://localhost:1234
