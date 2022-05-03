# !/bin/bash
./fetch_tasks.sh

python3 fill_db.py

if $WEB_HTTP; then
    DOMAIN=http://$WEB_HOST:$WEB_PORT/
else
    DOMAIN=https://$WEB_DOMAIN/
fi
echo "Testing $DOMAIN"

echo "Register: $TEST_MAIL
$(curl -s -X POST -F email=$TEST_MAIL -F pass=$TEST_PASS -F name=$TEST_NAME ${DOMAIN}register)"

echo "Restore: $TEST_MAIL
$(curl -s -X POST -F email=$TEST_MAIL -F pass=${TEST_PASS}_new ${DOMAIN}restore)"

TOKEN=$(curl -s ${DOMAIN}login?email=$TEST_MAIL\&pass=${TEST_PASS}_new | jq '.["token"]' | tr -d '"')
echo "Token: $TOKEN"

echo "Profile: $(curl -s ${DOMAIN}profile?token=$TOKEN)"

TASKS=$(curl -s -X GET ${DOMAIN}tasks/hierarchy?token=$TOKEN\&folders=sample_tests,unit-2,task-1)
echo "Existing tasks: $TASKS"

TASK_ID=$(echo $TASKS | jq '.sample_tests["units"]["unit-2"]["tasks"]["task-1"]["id"]')
echo "Test task $TASK_ID"

echo "Languages: $(curl -s ${DOMAIN}languages)"

VERBOSE='false'
send-solution() {
    RESP=$(curl -s -X POST ${DOMAIN}solution?token=$TOKEN \
        -F task_id=$TASK_ID \
        -F lang=$2 \
        -F verbose=$VERBOSE \
        --form-string source_text="$3")
    WAIT=$(echo $RESP | jq '.["wait"]')
    if [[ $WAIT != 'null' ]]; then
        echo "[...] Waiting for $WAIT to satisfy call limits"
        sleep $WAIT
        send-solution $1 $2 "$3"
    else
        echo "[$1] Response solution in $2: $RESP
        "
    fi
}

send-solution 'normal' 'c' '#include <stdio.h>
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}
'
send-solution 'normal' 'py' 'print(int(input())+int(input()))'
send-solution 'normal' 'pas' 'var
    A, B: Integer;
begin
    Read(A);
    Read(B);
    writeln(A+B);
end.'

send-solution 'wrong' 'c' '
#include <stdio.h>
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",1+b);}'
send-solution 'wrong' 'py' 'res = int(input())+1
print(res)'
send-solution 'wrong' 'pas' 'var
    A, B: Integer;
begin
    Read(A);
    Read(B);
    writeln(A+B-10);
end.'

send-solution 'hacky' 'c' '#include <stdio.h>
#include <stdlib.h>
int main(){system("ls");}'
send-solution 'hacky' 'py' '__builtins__.__import__("os").system("env")'
send-solution 'hacky' 'pas' '{$mode objfpc}{$H+}'"
uses 
  Process;
var 
    s : ansistring;
begin
    if RunCommand('/bin/bash',['cd / && ls'],s) then
           writeln(s); 
end."

send-solution 'malformed' 'c' '#include <stdio.h>
int main(){nt a,b;canf("%d%d",&a,&b);printf("%d",a+1+b);}'
send-solution 'malformed' 'py' 'res = int(input()))+int(input)
print(res)'
send-solution 'malformed' 'pas' 'var
var
    A, B: Integer;
beg
    Rad(A);
    Read(B);
    writeln(AB-10);
end.'

VERBOSE='true'
send-solution 'malformed' 'c' '#include <stdio.h>
int main(){nt a,b;canf("%d%d",&a,&b);printf("%d",a+1+b);}'

echo "Get template for C: $(curl -s ${DOMAIN}template?token=$TOKEN\&lang='c')"

echo "Leaderboard: $(curl -s ${DOMAIN}leaderboard?token=$TOKEN)"
