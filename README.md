# Language Agnostic TEsting

Web service that allows to run tests for programms written in these languages:

* C
* Python
* Planned: Go, C++, C#, Pascal

# How it works

* ✉️ Web service receives solution source code for specific task
* 🔨 Source code is built inside separate docker container
* 🧪 If build succeeded, then solution is tested with various test cases
* 📊 User receives test result

# Requirements

* docker-compose

# Architecture

Service have 4 containers:

* 🕸 web - web service written in Go, that:
	* Receives requests from clients
	* Communicates with **db**
	* Sends solutions into runner container
	* Responds with test result
* 🏃 runner - internal web service written in Python, that:
	* Receives solutions from **web** service
	* Builds solutions (if it's not written with interpreted language)
	* Tests solutions
	* Responds with test result to **web** service
* 🏗 manage - container with Bash and Python script, that could be used for:
	* Filling database with tests
	* Creating users
	* Giving tokens to users, that's required to send any solutions for testing
* 🗄 db - PostgreSQL container (postgres:latest)

# How to use

## TLDR; Commands for first start

```bash
git clone git@github.com:kee-reel/LATE.git late # Clone this repo
cd late # Go inside

sudo docker-compose up -d # Run all containers in detached mode

# Get id of manage container and open bash inside "manage" of it
sudo docker exec -it $(sudo docker ps | grep late_manage | cut -d' ' -f1) bash
```

Inside **manage** container:

```bash
# Stage 1 - preparing tests
mkdir tests # Create tests folder inside project directory
cd tests # Go inside
git clone https://github.com/kee-reel/late-sample-project # Clone sample project
cd .. # Go back

# Stage 2 - creating user and giving away token
python3 fill_db.py # Fill database with sample project
python3 create_user.py TestUser test@email.com # Create new user for testing
python3 create_token.py test@email.com # Give token for test user

# Stage 3 - usage (same code could be find in test_send.sh and test_post.sh)

# GET example - get all available tasks
curl http://web:1234?token=$(python3 get_test_token.py) # Send GET request 

# POST example - test solution file
curl -F token=$(python3 get_test_token.py) -F source_file=@tests/late-sample-project/unit-1/task-1/complete_solution.c -F task_id=3 -F verbose=true http://web:1234

# POST example - test solution text (this solution won't pass tests)
curl -F token=$(python3 get_test_token.py) -F source_text='print(4)' -F task_id=1 -F verbose=true http://web:1234
```

## Tests structure

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
	* "output" - text description of output format

Apart from `desc.json` file, task folder also must contain 2 files:

* `complete_solution.[c|py]` - file with source code of reference solution. Output of this file will be compared with incoming solutions - if output differs, than test of incoming solution fails
* `fixed_tests.txt` - file with tests for solution. It contains values that will be passed into both reference and incoming solutions

I have [repository](https://github.com/kee-reel/late-sample-project) with example project - you can use it for for reference.

## Service start

You can easily start web service with docker-compose:

```
$ docker-compose up # Add "-d" to run it in detached mode
```

After that you can manage web server via **manage** container. To open interactive bash shell inside of **manage** run:

```
# Get id of manage container and open bash inside "manage" of it
sudo docker exec -it $(sudo docker ps | grep late_manage | cut -d' ' -f1) bash
```

Inside this container you can run this scripts, to fammiliarize yourself with general workflow:

```
mkdir tests # Create tests folder inside project directory
cd tests # Go inside
git clone git@github.com:kee-reel/late-sample-project.git # Clone sample project
cd .. # Go back

python3 fill_db.py # Fill database with sample project
./create_admin.sh # Create admin user for testing
./test_get.sh # Get project's data
./test_post.sh 1 tests/late-sample-project/unit-1/task-1/complete_solution.c # Post solution
```

Scripts description:

* `fill_db.py` - fills database with tasks, contained inside "tasks" folder
* `create_admin.sh` - creates new user in database and gives token, that will be used to send solutions for tasks
* `test_get.sh` - get tasks currently present inside database
* `test_post.sh TASK_ID SOLUTION_FILE_PATH` - send solution for specific task (task id you can get from previos script)
* `add_user.py NICK EMAIL` - add user into database
* `create_token.py EMAIL PROJECT_ID` - give token to user

