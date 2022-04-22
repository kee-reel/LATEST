ARG IMAGE
FROM $IMAGE

USER root
# Some images doesn't create docker user
RUN if [ ! -e /home/docker ]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

COPY requirements.txt .

ENV TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apt update #&& apt upgrade -y
RUN apt install -y fpc 
RUN apt install -y python3
RUN apt install -y pip

USER docker
ENV PATH $PATH:/home/docker/.local/bin
RUN python3 -m pip install --upgrade pip
RUN pip3 install -r requirements.txt

USER root
COPY . .
RUN mkdir $DIRPATH/uploads && chmod 777 $DIRPATH/uploads

USER docker
CMD ["python3", "main.py"]
