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

# Setup environment

```bash
# Run all containers in detached mode for dev environment
./run-docker-compose.sh dev up -d

# Get id of "manage" container and open interactive bash shell inside of it
sudo docker exec -it $(sudo docker ps | grep late_manage | cut -d' ' -f1) bash
```

Inside **manage** container:

```bash
# Fill database with sample project, create user and run all available requests
./test_service.sh
```

# Requests

## Registration/login

Get user token, that will be used in all other requests. New user will be created if "email" is unknown to server.

```bash
curl -X GET https://DOMAIN/login?email=test@test.com\&pass=123456
```

### Query parameters:

* email - login
* pass - password (must be at least 6 symbols)

### Result example:

```json
{"token":"MzWNRaVruqAMbq60g0TqkFVFeFLnW9ECgThSSIo5XoFBUlCw6tzHElSqxhV8P8F24w25yTlUHPpttJanfbsKaH2NMKVR1yu8YCm6nfstbNLcXCbQSfW6LowfeDoERJGwuEQr2UKJVYlBCzN9an5ndxPucz4sxWbEmAqbsNM38eAqHcQYjQqdu0icjwI7h9fi8CNSPTECzvxFbeeq9EonZgMTLmmXkWqb4I9wLupT80Avy3kQ6Xxkp9thcMLIRP9i"}
```

## Get available tasks

Returns data about all projects, units and tasks stored in database. To send solutions you need to pick id (key in "tasks") for according task.

### Query parameters:

* token - token, returned on registration/login

```bash
curl -X GET https://DOMAIN?token=MzWNRaVruqAMbq60g0TqkFVFeFLnW9ECgThSSIo5XoFBUlCw6tzHElSqxhV8P8F24w25yTlUHPpttJanfbsKaH2NMKVR1yu8YCm6nfstbNLcXCbQSfW6LowfeDoERJGwuEQr2UKJVYlBCzN9an5ndxPucz4sxWbEmAqbsNM38eAqHcQYjQqdu0icjwI7h9fi8CNSPTECzvxFbeeq9EonZgMTLmmXkWqb4I9wLupT80Avy3kQ6Xxkp9thcMLIRP9i
```

### Result example:

> is\_passed - is given tasks was passed by user.
> Meaning of other fields could be found in chapter "Tests structure".

```json
{
   "projects": {
      "1": {
         "name": "Competition"
      }
   },
   "tasks": {
      "1": {
         "desc": "Сложить два числа и вывести результат",
         "input": [
            {
               "dimensions": [
                  1
               ],
               "name": "A",
               "range": [
                  "-1000",
                  "1000"
               ],
               "type": "int"
            },
            {
               "dimensions": [
                  1
               ],
               "name": "B",
               "range": [
                  "-1000",
                  "1000"
               ],
               "type": "int"
            }
         ],
         "is_passed": false,
         "language": "py",
         "name": "Сложение",
         "number": 0,
         "output": "Результат сложения A и B",
         "project": 1,
         "unit": 1
      },
      "2": {
         "desc": "Вывести строку \"Hello world!\"",
         "input": [],
         "is_passed": false,
         "language": "c",
         "name": "Hello world",
         "number": 0,
         "output": "Строка \"Hello world!\"",
         "project": 1,
         "unit": 2
      },
      "3": {
         "desc": "На вход даётся N чисел. Сложить между собой нечётные, вычесть из них чётные и вывести результат. Сначала на вход подаётся количество чисел, а затем сами числа.",
         "input": [
            {
               "dimensions": [
                  50
               ],
               "name": "A",
               "range": [
                  "-1000",
                  "1000"
               ],
               "type": "int"
            }
         ],
         "is_passed": false,
         "language": "c",
         "name": "Сложить нечётные, вычесть чётные",
         "number": 0,
         "output": "Результат сложения и вычитания чисел",
         "project": 1,
         "unit": 3
      }
   },
   "units": {
      "1": {
         "name": "Intro"
      },
      "2": {
         "name": "Intro"
      },
      "3": {
         "name": "Intro"
      }
   }
}
```

## Send solution to testing

Sends solution for specified task.

### Query parameters:

* token - token, returned on registration/login

### Form fields:

* task\_id - id of task
* source\_text - text of task solution
* source\_file - file with task solution
* verbose - expanded testing data will be returned ("false" by default)

> Either source\_text or source\_file must be specified

```bash
curl -X POST https://DOMAIN?token=MzWNRaVruqAMbq60g0TqkFVFeFLnW9ECgThSSIo5XoFBUlCw6tzHElSqxhV8P8F24w25yTlUHPpttJanfbsKaH2NMKVR1yu8YCm6nfstbNLcXCbQSfW6LowfeDoERJGwuEQr2UKJVYlBCzN9an5ndxPucz4sxWbEmAqbsNM38eAqHcQYjQqdu0icjwI7h9fi8CNSPTECzvxFbeeq9EonZgMTLmmXkWqb4I9wLupT80Avy3kQ6Xxkp9thcMLIRP9i \
	-F task_id=1 \
	--form-string source_text='#include <stdio.h>
int main(){int a,b;scanf("%d%d",&a,&b);printf("%d",a+b);}' \
	-F verbose=false
```

### Result example - no errors:

```json
{"error":null}
```

### Result example - build error:

```json
{"error":{"stage":"build","msg":"./uploads/solution_oK0y2aFu.c: In function ‘main’:\n./uploads/solution_oK0y2aFu.c:6:2: warning: implicit declaration of function ‘rintf’ [-Wimplicit-function-declaration]\n    6 |  rintf(\"%d\", a + 1 + b);\n      |  ^~~~~\n./uploads/solution_oK0y2aFu.c:6:2: warning: incompatible implicit declaration of built-in function ‘rintf’\n./uploads/solution_oK0y2aFu.c:2:1: note: include ‘<math.h>’ or provide a declaration of ‘rintf’\n    1 | #include <stdio.h>\n  +++ |+#include <math.h>\n    2 | int main()\n./uploads/solution_oK0y2aFu.c:6:8: error: incompatible type for argument 1 of ‘rintf’\n    6 |  rintf(\"%d\", a + 1 + b);\n      |        ^~~~\n      |        |\n      |        char *\n./uploads/solution_oK0y2aFu.c:6:8: note: expected ‘float’ but argument is of type ‘char *’\n./uploads/solution_oK0y2aFu.c:6:2: error: too many arguments to function ‘rintf’\n    6 |  rintf(\"%d\", a + 1 + b);\n      |  ^~~~~\n"},"fail_count":0}
```

### Result example - test error:

```json
{"error":{"stage":"test","expected":"2","params":"1;1;","result":"3"},"fail_count":0}
```

### Result example - verbose mode (verbose parameter set to "true"):

```json
{"error":null,"fail_count":0,"result":[{"params":"1;1;","result":"2"},{"params":"0;0;","result":"0"},{"params":"-1;1;","result":"0"},{"params":"10;10;","result":"20"},{"params":"20;-20;","result":"0"},{"params":"-100;-100;","result":"-200"},{"params":"347;-379;","result":"-32"},{"params":"-313;137;","result":"-176"},{"params":"-319;491;","result":"172"},{"params":"268;-819;","result":"-551"},{"params":"-296;-546;","result":"-842"},{"params":"435;-123;","result":"312"},{"params":"878;-621;","result":"257"},{"params":"110;79;","result":"189"},{"params":"546;330;","result":"876"},{"params":"533;786;","result":"1319"},{"params":"-45;535;","result":"490"},{"params":"439;973;","result":"1412"},{"params":"-615;561;","result":"-54"},{"params":"-958;-703;","result":"-1661"},{"params":"855;-408;","result":"447"},{"params":"767;-154;","result":"613"},{"params":"-413;278;","result":"-135"},{"params":"-461;23;","result":"-438"},{"params":"-425;913;","result":"488"},{"params":"142;656;","result":"798"},{"params":"-53;-950;","result":"-1003"},{"params":"-539;814;","result":"275"},{"params":"-229;-918;","result":"-1147"},{"params":"-619;56;","result":"-563"},{"params":"-736;151;","result":"-585"},{"params":"407;102;","result":"509"},{"params":"-789;544;","result":"-245"},{"params":"-238;668;","result":"430"},{"params":"742;-848;","result":"-106"},{"params":"129;-207;","result":"-78"}]}
```

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
* 🏗 manage - container with Bash and Python scripts, that could be used for:
	* Filling database with tests
	* Creating users
	* Giving tokens to users, that's required to send any solutions for testing
* 🗄 db - PostgreSQL container (postgres:latest)

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

This is example of `desc.json` file for `task`:

```json
{
	"name": "Addition",
	"desc": "Add two numbers and output the result",
	"input": [
		{"name": "A", "type": "int", "range": ["-1000", "1000"]}, 
		{"name": "B", "type": "int", "range": ["-1000", "1000"]}
	],
	"output": "Result of adding A to B"
}
```

Apart from `desc.json` file, task folder also must contain 2 files:

* `complete_solution.[c|py]` - file with source code of reference solution. Output of this file will be compared with incoming solutions - if output differs, than test of incoming solution fails
* `fixed_tests.txt` - file with tests for solution. It contains values that will be passed into both reference and incoming solutions

I have [repository](https://github.com/kee-reel/late-sample-project) with example project - you can use it for for reference.

## Service start

You can easily start web service with docker-compose:

```bash
./run-docker-compose.sh dev up -d # Run all containers in detached mode for dev environment
```

After that you can manage web server via **manage** container. To open interactive bash shell inside of **manage** run:

```bash
# Get id of manage container and open bash inside "manage" of it
sudo docker exec -it $(sudo docker ps | grep late_manage | cut -d' ' -f1) bash
```

Then you need to prepare tests - you can use mine for this time:

```bash
mkdir tests # Create tests folder
cd tests # Go inside
git clone https://github.com/kee-reel/late-sample-project # Clone sample project
cd .. # Go back
```

Test are ready, lets insert them into database and create new user:


```bash
python3 fill_db.py # Fill database with sample project
```

All set, now we can try to send requests to web server. You can check request examples in [/manage/test\_service.sh](/manage/test_service.sh).
