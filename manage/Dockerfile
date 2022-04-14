ARG IMAGE
FROM ubuntu:latest

USER root
# Some images doesn't create docker user
RUN if [ ! -e /home/docker ]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

COPY requirements.txt .

ENV TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apt update && apt upgrade -y 
RUN apt install -y python3 python3-pip
RUN apt install -y libpq-dev
RUN apt install -y jq curl git

USER docker

ENV PATH $PATH:/home/docker/.local/bin
RUN pip3 install -r requirements.txt

COPY . .
