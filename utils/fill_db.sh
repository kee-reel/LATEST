for f in $(find tasks -iname desc.json); do
	echo $f
	python3 utils/insert_task_to_db.py $f
done
