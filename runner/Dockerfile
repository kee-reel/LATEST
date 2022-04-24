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
# Start
USER $USER
CMD ["python3", "main.py"]
