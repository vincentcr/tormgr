#!/bin/sh

FORCE_AUTH="host all  all 0.0.0.0/0 md5"
sed -i -e "s|host *all *all *0\.0\.0\.0/0 *trust|${FORCE_AUTH}|g" /var/lib/postgresql/data/pg_hba.conf
echo "${FORCE_AUTH}" >> /var/lib/postgresql/data/pg_hba.conf
