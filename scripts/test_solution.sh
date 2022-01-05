#!/bin/bash
TIMEOUT=3

run_tests () {
	fix_file $1
	cat $1 | while read test_case; do
		if [[ -z "$test_case" ]]; then
			continue
		fi
		actual=$(echo "${test_case};" | tr ';' '\n' | timeout $TIMEOUT runuser -u testrun ./solution | tail -1)
		if [[ $? -ne 0 ]]; then
			exit $?
		fi
		target=$(echo "${test_case};" | tr ';' '\n' | timeout $TIMEOUT ./complete_solution | tail -1)
		if [[ $? -ne 0 ]]; then
			exit $?
		fi
		if [[ "$actual" == "$target" ]]; then
			echo "Параметры: $test_case
Результат:
$(echo $actual | sed 's/||/\n/g')
" >> test_result.txt
		else
			echo "	Параметры: $test_case
Ожидалось:
$(echo $target | sed 's/||/\n/g')
Получено:
$(echo $actual | sed 's/||/\n/g')
" > /dev/stderr
			exit -1
		fi
	done
}

fix_file() {
	# If file have no newline at the end "read" won't read it's last line
	if [[ -n "$(tail -1 $1)" ]]; then
		echo >> user_tests.txt
	fi
}

cd $1

> test_result.txt

echo "1. Пользовательские тесты:
" >> test_result.txt &&
run_tests user_tests.txt &&

echo '+' &&

echo "2. Фиксированные тесты:
" >> test_result.txt &&
run_tests fixed_tests.txt &&

echo "3. Рандомные тесты:
" >> test_result.txt &&
run_tests random_tests.txt
