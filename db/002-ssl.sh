#!/bin/sh

CWD=$(pwd)
SSL_DIR=$PGDATA/ssl
PGCONF=$PGDATA/postgresql.conf

# Install self-signed certificate and disallow non-SSL connections
mkdir -p $SSL_DIR && cd $SSL_DIR && \
    openssl req -new -newkey rsa:1024 -days 365000 -nodes -x509 \
      -keyout server.key -subj "/CN=PostgreSQL" -out server.crt && \
    chmod og-rwx server.key && chown -R postgres:postgres $SSL_DIR

cat <<EOT >> ${PGCONF}
ssl = on
ssl_ciphers = 'DEFAULT:!LOW:!EXP:!MD5:@STRENGTH'
ssl_cert_file = '${SSL_DIR}/server.crt'
ssl_key_file = '${SSL_DIR}/server.key'
EOT

cd ${CWD}
