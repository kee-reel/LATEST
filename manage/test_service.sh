# !/bin/bash

if [[ -z "$(ls tests)" ]]; then
	cd tests # Go inside
	git clone https://github.com/kee-reel/late-sample-project # Clone sample project
	cd ..
	python3 fill_db.py
fi

DOMAIN=http$(if [[ "$WEB_HTTP" == 'false' ]]; then echo s; fi)://$WEB_HOST:$WEB_PORT$WEB_ENTRY
echo "Testing $DOMAIN"

TOKEN=$(curl -s ${DOMAIN}login?email=test@test.com\&pass=123456 | grep -Po '[\w\d]{256}')

echo "Token: $TOKEN
===
Existing tasks:"
TASKS=$(curl -X GET $DOMAIN?token=$TOKEN)
echo $TASKS

TASK_ID=$(echo $TASKS | jq '.tasks | keys[1]' | tr -d '"')
echo "Test task $TASK_ID"

echo '
===
Post solution:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=false

echo '
===
Post wrong solution:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+1+b);}' \
	-F verbose=false

echo '
===
Post malformed solution:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
	--form-string source_text='#include <stdio.h> 
int main(){nt a,b;canf("%d%d",&a,&b);printf("%d",a+1+b);}' \
	-F verbose=false

echo '
===
Post solution with verbose flag:'
curl -X POST $DOMAIN?token=$TOKEN \
	-F task_id=$TASK_ID \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=true
