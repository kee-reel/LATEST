FROM redis:alpine

ARG REDIS_RESULT_LIST
ARG REDIS_TASK_LIST
ARG REDIS_RUNNER_USER
ARG REDIS_RUNNER_PASSWORD

ENV ACL_FILE /etc/redis/users.acl
COPY users.acl $ACL_FILE
RUN sed -i "s/REDIS_RESULT_LIST/$REDIS_RESULT_LIST/g" $ACL_FILE && \
    sed -i "s/REDIS_TASK_LIST/$REDIS_TASK_LIST/g" $ACL_FILE && \
    sed -i "s/REDIS_RUNNER_USER/$REDIS_RUNNER_USER/g" $ACL_FILE && \
    sed -i "s/REDIS_RUNNER_PASSWORD/$REDIS_RUNNER_PASSWORD/g" $ACL_FILE

ENV CONF_FILE /usr/local/etc/redis/redis.conf
COPY redis.conf $CONF_FILE

CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]
