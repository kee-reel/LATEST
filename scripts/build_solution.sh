#!/bin/bash
P=$1
EXT=$2

if [[ $EXT != 'c' ]]; then
	echo -n "$EXT"
	exit
fi

cd $P
rm -f complete_solution.exe
/usr/bin/gcc -o complete_solution.exe complete_solution.c -lm
rm -f solution.exe
/usr/bin/gcc -o solution.exe solution.c -lm

echo -n 'exe'
