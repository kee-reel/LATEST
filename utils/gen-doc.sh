# !/bin/bash

if [[ ! -e utils/swag ]]; then
    SWAG_VER='1.8.1'
    SWAG_FILE="swag_${SWAG_VER}_Linux_$(uname -m)"
    mkdir utils/.swag
    cd utils/.swag
    wget https://github.com/swaggo/swag/releases/download/v$SWAG_VER/$SWAG_FILE.tar.gz
    tar -xvzf $SWAG_FILE.tar.gz
    cd ../..
    mv utils/.swag/swag utils/swag
    rm -rf utils/.swag
fi
cd web
../utils/swag init
mv docs/swagger.json .
rm -rf docs
