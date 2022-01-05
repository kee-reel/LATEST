Language AGNOstic Testing

Initially, I wrote this web-service to run autotests of C language labs solutions from students.

In result I've got service that can autotest any program that could receive standard input and produce standard output.

# How testing works

Testing of solution have following stages:

* Receiving solution file
* Compilation of solution file
* Generation of random test cases
* Execution of reference solution with:
	* User test cases
	* Fixed test cases
	* Random test cases
* Execution of received solution with same test cases
* Comparison of standard output of reference and received solution
* Production of test report

![Test result](/res/test-result.jpg)

# How to start service

Start application with:
```
go run .
```
After that stop application.

On start, application creates tasks.db file, that stores all neccessary data.

To add new lab tasks:
* Open "tasks" folder
* Add new subject folder in format "subject-NUMBER", where NUMBER is integer
* Inside subject folder add new work folder in format "work-NUMBER"
* Inside work folder add new variant folder in format "variant-NUMBER"
* Inside variant folder add new task folder in format "task-NUMBER"
* Inside task folder create files "desc.json", "complete_solution.c" and "fixed_tests.c"
* Run scripts/fill_db.sh script, to automatically fill tasks.db with created tasks

Scripts for DB management:
* scripts/fill_db.sh -- fills tasks.db with complete solutions from "tasks/" folder
* scripts/fill_users.sh FILENAME -- fills tasks.db with users. Each line in file must be in this format: "name last_name group_name". Group name must be in format like this: "o717b01"
* python3 scripts/create_token.py SUBJECT WORK -- fills tasks.db with access tokens for specified SUBJECT and WORK. SUBJECT and WORK is numbers in name of respective tasks folder

You can check how autotests is built in "scripts/build_solution.sh" and how it is ran in "scripts/test_solution.sh".
