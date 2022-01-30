# !/bin/bash
if (( $# == 0 )); then
	echo "Usage: $0 [dev | prod] commands"
	exit
fi

ARGS=($@)
E=$1
sudo docker-compose -f docker-compose.yml -f docker-compose.$(arch).yml -f docker-compose.$E.yml ${ARGS[@]:1}
