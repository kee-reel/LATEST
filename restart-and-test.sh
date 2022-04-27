if (( $# == 0 )); then
    ./run.sh up --build -d
else
    ./run.sh up --build -d --force-recreate --no-deps $@
fi
sudo docker exec $(sudo docker ps | grep manage | cut -d' ' -f1) bash -c './test_service.sh'
