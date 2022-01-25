# Language Agnostic TEsting

Web service that allows to run tests for programms written in these languages:

* C
* Python
* Planned: Go, C++, C#, Pascal

### How testing works

* ✉️ Web service receives solution source code for specific task
* 🔨 Source code is built inside separate docker container
* 🧪 If build succeeded, then solution is tested with various test cases
* 📊 User receives test result

### Service containers

Service have 3 containers:

* 🕸 web - web service written in Go, that:
	* Receives requests
	* Communicates with **db**
	* Sends solutions into runner container
	* Responds with test result
* 🏃 runner - internal web service written in Python, that:
	* Receives solutions from **web** service
	* Builds solutions (if it's not written with interpreted language)
	* Tests solutions
	* Responds with test result
* 🗄 db - PostgreSQL container (postgres:latest)

### How to start service

You can easily start whole web service with docker-compose:

```
$ docker-compose up
```

### How to add tasks for testing

> To manage web service you need to have Bash and Python3 installed.

Folder "utils" contains various scripts written with Bash or Python - these scripts implements various management functionallity.

On first run I recommend to run these scripts:

* `fill_db.sh` - fills database with tasks, contained inside "tasks" folder
* `create_admin.sh` - creates new user in database and gives token, that will be used to send solutions for tasks
* `test_get.sh` - get tasks currently present inside database
* `test_send.sh TASK_ID SOLUTION_FILE_PATH` - send solution for specific task (task id you can get from previos script)
