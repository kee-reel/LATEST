ARG IMAGE
FROM $IMAGE

USER root
# Some images doesn't create docker user
RUN if [ ! -e /home/docker ]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

COPY requirements.txt .

RUN apt update && apt upgrade -y 
RUN apt install -y libpq-dev python3 python3-pip

USER docker

ENV PATH $PATH:/home/docker/.local/bin
RUN pip3 install -r requirements.txt

COPY . .
