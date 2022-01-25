#!/bin/bash
TASK_ID=$1
TASK_FILE=$2
T=$(python3 utils/get_test_token.py)
curl -F token=$T -F source_${TASK_ID}=@$(pwd)/$TASK_FILE -F tasks=$TASK_ID -F verbose=true http://localhost:1234
