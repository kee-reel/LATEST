# !/bin/bash

if [ -e tests/$GIT_REPO_FOLDER ]; then
	cd tests/$GIT_REPO_FOLDER
	git pull
	cd ../..
elif [ -n $GIT_REPO ]; then
	git clone $GIT_REPO tests/$GIT_REPO_FOLDER
else
	echo 'Git repo is not specified'
	exit
fi
python3 fill_db.py
