cat $1 | while read s; do
	data=($(echo $s | tr ';' ' '))
	python3 scripts/add_user.py  ${data[@]}
done
