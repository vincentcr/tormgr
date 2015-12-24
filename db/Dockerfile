FROM postgres:9.4

ENV APP_NAME=tormgr
ENV DB_NAME=${APP_NAME}_dev
ENV DB_USER=${APP_NAME}_dev
ENV DB_PASSWD=${APP_NAME}_dev_very_secret
ENV SQL_DIR=/docker-entrypoint-initdb.d

RUN mkdir -p $SQL_DIR
WORKDIR $SQL_DIR
COPY . $SQL_DIR
