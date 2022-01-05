#!/bin/bash
cd $1
rm -f complete_solution
/usr/bin/gcc -o complete_solution -I../../../.. complete_solution.c -lm
rm -f solution
/usr/bin/gcc -o solution -I../../../.. solution.c -lm
