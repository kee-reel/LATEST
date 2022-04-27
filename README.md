# Language Agnostic TESTing

Main goal of this project is to provide web-service that allows teachers to create programming courses with built-in interactive excercises.

This web-service supports tests for programs written in these languages:

* C
* C++
* Python
* Pascal
* Planned: Go, C#

# Contents
- [How it works](#how-it-works)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Architecture](#architecture)
- [Tests structure](#tests-structure)
- [Service start](#service-start)
- [Requests](#requests)

# How it works

* ✉️ Web service receives solution source code for specific task
* 🔨 Source code is built inside separate docker container
* 🧪 If build succeeded, then solution is tested with various test cases
* 📊 User receives test result

This testing system is "language agnostic" because:

* All input parameters is passed via standard input
* Result is received in specific format via standard output
* Teacher provides task solution only in one language
* Standard output of all students' solutions will be compared against teacher's solution

This is example of program for this testing system:

```python
n = int(input()) # Receive count
s = 0
for _ in range(n):
    s += int(input()) # Receive numbers n times
print(s) # Output addition result
```

> Yes I know about command line arguments, but I've built it this way, so programms still can be executed and tested manually as usual.
> 
> I don't consider that it's good idea to teach begginers in programming about command line arguments at first lesson.

# Requirements

* docker-compose

# Quick start

```bash
# Run all containers in detached mode for dev environment
./run-docker-compose.sh dev up -d

# Get id of "manage" container and open interactive bash shell inside of it
sudo docker exec -it $(sudo docker ps | grep manage | cut -d' ' -f1) bash
```

Inside opened bash shell:

```bash
# Fill database with sample project, create user and run all available requests
./test_service.sh
```

# Architecture

Service have 4 containers:

* 🕸 web - web service written in Go, that:
	* Receives request from clients
	* Authenticates client using access token stored inside **redis**
	* Manages user and task related data stored inside **db**
	* Puts solution and takes results from **redis** queue 
	* Responds with test result to client
* 🏃 runner - internal web service written in Python, that:
	* Takes solution from **redis** job queue
	* Builds solution
	* Tests solution
	* Sends test result to **redis** result queue
* 🏗 manage - container with Bash and Python scripts, that is used for:
	* Filling **db** with tasks
	* Testing **web** service
* 🗄 db - PostgreSQL container (postgres:latest)
* 📋 redis - Redis container (redis:alpine)

# Tests structure

Main purpose of this web service is testing of specific programms, so let's figure out how you need to set them up.

Tests is organized this way:

`"tests"` -> `project` -> `unit` -> `task`

* `"tests"` - folder in project root directory, that contains projects
* `project` - folder with arbitrary name, that contains units
* `unit` - folder with arbitrary name, that contains tasks
* `task` - folder with arbitrary name, that contains actual test data

`project`, `unit` and `task` folders contains file `desc.json`, that contains descripton for according folder. Here are neccessary fields for every folder type:

* `project`
	* "name" - human readable name of project
* `unit`
	* "name" - human readable name of unit
* `task`
	* "name" - human readable name of project
	* "position" - position inside unit when it will be presented to user
	* "desc" - text description that will help user to understant given task
	* "input" - format of input data for program
		* "name" - name of input parameter
		* "type" - type of passed values
		* "range" - range (from min to max) in which values for random tests will be generated
		* "dimensions" - if this field is not specified, then it is just single value, if value specified - it is specification of array size and dimensions. Each value specifies size of dimension. Examples:
			* [50] - programm could receive array from 1 up to 50 elements (size will be randomly generated in each test)
			* [10, 5] - matrix 10x5 (each size for each dimension will be generated randomly in range [1,10]x[1,5])
			* [3, 0] - if zero is specified, size will be the same as previous one (in given example, possible sizes for matrix are 1x1, 2x2, 3x3)
	* "output" - text description of output format

This is example of `desc.json` file for some `task`:

```json
{
	"name": "Add to array",
	"desc": "Add a number to all values in an array",
	"input": [
		{"name": "A", "type": "int", "range": ["-1000", "1000"]}, 
		{"name": "B", "type": "int", "range": ["-1000", "1000"], "dimensions": [50]}
	],
	"output": "Result of adding A to B"
}
```

Apart from `desc.json` file, task folder also must contain other files:

* `complete_solution.*` - file with source code of reference solution. Output of this file will be compared with incoming solutions - if output differs, than test of incoming solution fails
* `fixed_tests.txt` - file with tests for solution. It contains values that will be passed into both reference and incoming solutions
* `template.*` - file with template for solution. Contents of this file could be used on UI side, to provide user with sample code for easy start

I have [repository](https://github.com/kee-reel/latest-sample-project) with example project - you can use it for for reference.

# Service start

You can easily start web service with docker-compose:

```bash
./run-docker-compose.sh dev up -d # Run all containers in detached mode for dev environment
```

> You can check file [/manage/test\_service.sh](/manage/test_service.sh) - it sets up environment and makes sample requests.

After that you can manage web server via **manage** container. To open interactive bash shell inside **manage**:

```bash
sudo docker exec -it $(sudo docker ps | grep manage | cut -d' ' -f1) bash
```

Then you need to fetch tasks that you want to insert in **db**:

```bash
./fetch_tasks.sh
```

Tasks are ready, lets insert them into **db**:

```bash
python3 fill_db.py # Fill database with sample project
```

All set, now we can try to send requests to web server by yourself or test server with script:

```bash
./test_service.sh
```

# Requests

API responds with error codes that ould be found in [web/api/errors.go](/web/api/errors.go)

If you know what Swagger is -- you can check specs for all requests in [swagger.json](/web/docs/swagger.json).

If you don't -- check out hosted Swagger docs [here](https://kee-reel.com/cyber-api/).
