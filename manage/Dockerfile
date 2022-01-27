ARG IMAGE
FROM $IMAGE
WORKDIR /app

USER root
RUN if [[ ! -e /home/docker ]]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

COPY requirements.txt .

USER docker

ENV PATH $PATH:/home/docker/.local/bin
RUN python3 -m pip install --upgrade pip
RUN pip install -r requirements.txt

COPY . .
