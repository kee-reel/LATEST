FROM matthewfeickert/docker-python3-ubuntu
WORKDIR /app

USER ${uid}:${gid}

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .
CMD ["bash"]
