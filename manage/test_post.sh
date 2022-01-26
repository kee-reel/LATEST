#!/bin/bash
TASK_ID=$1
TASK_FILE=$2
T=$(python3 get_test_token.py)
curl \
	-F token=$T \
	-F source=@$(pwd)/$TASK_FILE \
	-F task_id=$TASK_ID \
	-F verbose=true \
http://web:1234
