if (( $# == 0 )); then
    ./run-docker-compose.sh up --build -d
else
    ./run-docker-compose.sh up --build -d --force-recreate --no-deps $@
fi
sudo docker exec $(sudo docker ps | grep manage | cut -d' ' -f1) bash -c './test_service.sh'
