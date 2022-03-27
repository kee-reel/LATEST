# !/bin/bash

if [[ -z "$(ls tests)" ]]; then
	cd tests # Go inside
	git clone $GIT_REPO $GIT_REPO_FOLDER # Clone sample project
	git clone $TEMPLATES_GIT_REPO templates # Clone templates
	cd ..
	python3 fill_db.py
fi

if $WEB_HTTP; then
    DOMAIN=http://$WEB_HOST:$WEB_PORT$WEB_ENTRY
else
    DOMAIN=https://$WEB_DOMAIN$WEB_ENTRY
fi
echo "Testing $DOMAIN"

TOKEN=$(curl -s ${DOMAIN}login?email=$TEST_MAIL\&pass=$TEST_PASS | grep -Po '[\w\d]{256}')

echo "Token: $TOKEN
===
Existing tasks:"
TASKS=$(curl -X GET $DOMAIN?token=$TOKEN\&folders=true\&task_folders=sample_tests,unit-2,task-1)
echo $TASKS

TASK_ID=$(echo $TASKS | jq '.sample_tests["units"]["unit-2"]["tasks"]["task-1"]["id"]')
echo "Test task $TASK_ID"

echo '
===
Post solution in C:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='c' \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}'

echo '
===
Post solution in Python:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='py' \
    --form-string source_text='print(int(input())+int(input()))'

echo '
===
Post wrong solution in C:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='c' \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+1+b);}'

echo '
===
Post wrong solution in Python:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='py' \
    --form-string source_text='res = int(input())+int(input())+1
print(res)'

echo '
===
Post malformed solution in C:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='c' \
	--form-string source_text='#include <stdio.h> 
int main(){nt a,b;canf("%d%d",&a,&b);printf("%d",a+1+b);}'

echo '
===
Post malformed solution in Python:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='py' \
    --form-string source_text='res = int(input()))+int(input)
print(res)'

echo '
===
Post solution with verbose flag:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
    -F lang='c' \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=true

echo '
===
Get template for C:'
curl ${DOMAIN}template?token=$TOKEN\&lang='c'
