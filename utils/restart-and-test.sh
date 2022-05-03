if (( $# == 0 )); then
    ./utils/run.sh up --build -d
else
    ./utils/run.sh up --build -d --force-recreate --no-deps $@
fi
./utils/test.sh
