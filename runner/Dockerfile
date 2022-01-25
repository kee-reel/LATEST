FROM matthewfeickert/docker-python3-ubuntu
WORKDIR /app
EXPOSE 1337

USER root

RUN mkdir /app/uploads
RUN chmod -R 777 /app/uploads

USER ${uid}:${gid}

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .
CMD ["python3", "main.py"]
