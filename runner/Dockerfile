ARG IMAGE
FROM $IMAGE
USER root

ENV USER appuser
RUN addgroup -S appgroup && adduser -S $USER -G appgroup

ENV DIRPATH /home/$USER/app
WORKDIR $DIRPATH
# Install requirements
USER root
COPY requirements.txt .
USER $USER
RUN pip3 install -r requirements.txt
# Copy files and create directory
USER root
RUN apk update && apk add gcc musl-dev
RUN mkdir $DIRPATH/uploads && chmod 777 $DIRPATH/uploads
COPY . .

ENV FPC_VERSION="3.2.2" \
    FPC_ARCH="x86_64-linux"

# Free Pascal installation
# Taken from https://github.com/Docker-Hub-frolvlad/docker-alpine-fpc/blob/master/Dockerfile
RUN apk add --no-cache binutils && \
    cd /tmp && \
    wget "https://downloads.sourceforge.net/project/freepascal/Linux/${FPC_VERSION}/fpc-${FPC_VERSION}.${FPC_ARCH}.tar" -O fpc.tar && \
    tar xf "fpc.tar" && \
    cd "fpc-${FPC_VERSION}.${FPC_ARCH}" && \
    rm demo* doc* && \
    \
# Workaround musl vs glibc entrypoint for `fpcmkcfg`
    mkdir /lib64 && \
        ln -s /lib/ld-musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && \
        \
        echo -e '/usr\nN\nN\nN\n' | sh ./install.sh && \
        find "/usr/lib/fpc/${FPC_VERSION}/units/${FPC_ARCH}/" -type d -mindepth 1 -maxdepth 1 \
        -not -name 'fcl-base' \
        -not -name 'rtl' \
        -not -name 'rtl-console' \
        -not -name 'rtl-objpas' \
        -exec rm -r {} \; && \
        rm -r "/lib64" "/tmp/"*

# Start
USER $USER
CMD ["python3", "-u", "main.py"]
