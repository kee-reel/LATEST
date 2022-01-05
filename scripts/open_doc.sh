#!/bin/bash
json=$(</dev/stdin)
if [[ $json =~ pdf ]]; then
	cd ..
	okular $(echo $json | jq .link | tr -d '"')
else
	echo $json | sed 's/\\n/\n/g'
fi
