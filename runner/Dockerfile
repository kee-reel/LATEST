ARG IMAGE
FROM $IMAGE
EXPOSE 1337

USER root
RUN if [[ ! -e /home/docker ]]; then useradd -m docker; fi
ENV DIRPATH /home/docker/app
WORKDIR $DIRPATH

RUN mkdir $DIRPATH/uploads
RUN chmod -R 777 $DIRPATH/uploads

COPY requirements.txt .

USER docker
ENV PATH $PATH:/home/docker/.local/bin
RUN python3 -m pip install --upgrade pip
RUN pip3 install -r requirements.txt

COPY . .
CMD ["python3", "main.py"]
