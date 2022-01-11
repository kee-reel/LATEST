#!/bin/bash
TIMEOUT=3

run_tests () {
	if [[ ! -e $1 ]]; then
		exit 0
	fi
	F=$1
	PROMPT=$2
	fix_file $F
	echo \"$PROMPT\"': ['
	cat $F | while read test_case; do
		if [[ -z "$test_case" ]]; then
			continue
		fi
		actual=$(echo "${test_case};" | tr ';' '\n' | timeout $TIMEOUT  | tail -1)
		if [[ $? -ne 0 ]]; then
			exit $?
		fi
		target=$(echo "${test_case};" | tr ';' '\n' | timeout $TIMEOUT ./complete_solution | tail -1)
		if [[ $? -ne 0 ]]; then
			exit $?
		fi

		echo ",{\"params\": \"$test_case\", \"result\": \"$(echo $actual | sed 's/||/\n/g')\""
		if [[ "$actual" != "$target" ]]; then
			echo ",\"expected\": \"$(echo $actual | sed 's/||/\n/g')\""
		fi
		echo '}'
	done
	echo ']'
}

fix_file() {
	# If file have no newline at the end "read" won't read it's last line
	if [[ -n "$(tail -1 $1)" ]]; then
		echo >> user_tests.txt
	fi
}

cd $1

> test_result.txt

run_tests user_tests.txt 'user' >> test_result.txt &&
echo '+' &&
run_tests fixed_tests.txt 'fixed' >> test_result.txt &&
run_tests random_tests.txt 'random' >> test_result.txt
