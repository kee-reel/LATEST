# !/bin/bash
sudo docker-compose -f docker-compose.yml -f docker-compose.$(uname -m).yml $@
