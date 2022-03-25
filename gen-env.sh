# !/bin/bash

if [[ -e .env ]]; then
    echo '.env file already exists'
    exit
fi

echo 'POSTGRES_DB=tasks
POSTGRES_USER=root
POSTGRES_PASSWORD=pass
APP_DB_USER=docker
APP_DB_PASS=docker
APP_DB_NAME=docker

DB_HOST=db
DB_PORT=5432
DB_USER=root
DB_PASS=pass
DB_NAME=tasks

WEB_ENTRY=/
WEB_HOST=web
WEB_PORT=1234
WEB_HTTP=true
WEB_CERT_FILE=/etc/letsencrypt/live/YOURSITE/fullchain.pem
WEB_KEY_FILE=/etc/letsencrypt/live/YOURSITE/privkey.pem

MAIL_EMAIL=YOUR_EMAIL
MAIL_SERVER=SMTP_SERVER
MAIL_SERVER_PORT=SMTP_SERVER_PORT
MAIL_PASS=SMTP_SERVER_PASS
MAIL_SUBJECT=sample subject
MAIL_MSG=sample message

GIT_REPO=https://github.com/kee-reel/late-sample-project
GIT_REPO_FOLDER=sample_tests
TEMPLATES_GIT_REPO=https://github.com/kee-reel/late-solution-templates
TEST_MAIL=test@test
TEST_PASS=123456' > .env

echo 'Created .env file'
