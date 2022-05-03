sudo docker exec $(sudo docker ps | grep manage | cut -d' ' -f1) bash -c './test_service.sh'
