ARG IMAGE
FROM $IMAGE

USER root
# Some images doesn't create docker user
RUN if [ ! -e /home/docker ]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

COPY requirements.txt .

RUN apt update && apt upgrade -y 
RUN apt install -y python3 python3-pip libpq-dev curl git

RUN mkdir tests && chown docker tests

USER docker

ENV PATH $PATH:/home/docker/.local/bin
RUN pip3 install -r requirements.txt

COPY . .
