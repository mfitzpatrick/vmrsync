#! /bin/sh

cd "$(dirname $0)"

docker-compose up -d

sleep 5

docker-compose exec db /usr/local/firebird/bin/isql -u SYSDBA -i /setup/schema.sql
docker-compose exec db /usr/local/firebird/bin/isql -u SYSDBA -i /setup/initial-data.sql

echo "Firebird DB is ready for operations on port 3050"

