# !/bin/bash

if [[ -z "$(ls tests)" ]]; then
	cd tests # Go inside
	git clone https://github.com/kee-reel/late-sample-project # Clone sample project
	cd ..
	python3 fill_db.py
fi

DOMAIN=$WEB_HOST:$WEB_PORT$WEB_ENTRY
TOKEN=$(curl -s http://${DOMAIN}login?email=test@test.com\&pass=123456 | grep -Po '[\w\d]{256}')

echo "Token: $TOKEN
===
Existing tasks:"
curl -X GET http://$DOMAIN?token=$TOKEN

echo '
===
Post solution:'
curl -X POST http://$DOMAIN?token=$TOKEN \
	-F task_id=2 \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=false

echo '
===
Post wrong solution:'
curl -X POST http://$DOMAIN?token=$TOKEN \
	-F task_id=2 \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+1+b);}' \
	-F verbose=false

echo '
===
Post malformed solution:'
curl -X POST http://$DOMAIN?token=$TOKEN \
	-F task_id=2 \
	--form-string source_text='#include <stdio.h> 
int main(){nt a,b;canf("%d%d",&a,&b);printf("%d",a+1+b);}' \
	-F verbose=false

echo '
===
Post solution with verbose flag:'
curl -X POST http://$DOMAIN?token=$TOKEN \
	-F task_id=2 \
	--form-string source_text='#include <stdio.h> 
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=true
