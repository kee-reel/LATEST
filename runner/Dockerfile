ARG IMAGE
FROM $IMAGE

USER root
# Some images doesn't create docker user
RUN if [ ! -e /home/docker ]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

RUN mkdir $DIRPATH/uploads
RUN chmod -R 777 $DIRPATH/uploads

COPY requirements.txt .

RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC
RUN apt update && apt upgrade -y
RUN apt install -y fpc python3
RUN apt install -y pip

USER docker
ENV PATH $PATH:/home/docker/.local/bin
RUN python3 -m pip install --upgrade pip
RUN pip3 install -r requirements.txt

USER root
USER docker

COPY . .
CMD ["python3", "main.py"]
