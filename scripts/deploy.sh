#!/bin/bash
if [[ -z "$(pwd | grep -E 'labs-auto-test$')" ]]; then
	exit
fi
rsync -vaz ./* pi@kee-reel.com:/home/pi/labs-auto-test/
ssh pi@pi 'cd labs-auto-test; cp -r static/* ../blog/autotests/; cp -r static/* ../blog/_site/autotests/'
if [[ "$1" == '-r' ]]; then
	ssh pi@pi 'cd labs-auto-test; ./scripts/fill_db.sh;'
fi
