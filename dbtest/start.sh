#! /bin/sh

arg="$1"

cd "$(dirname $0)"

docker-compose up -d

sleep 5

if [ "$arg" = "test" ]; then
    docker-compose exec -T db /usr/local/firebird/bin/isql -u SYSDBA -i /setup/schema.sql
    docker-compose exec -T db /usr/local/firebird/bin/isql -u SYSDBA -i /setup/initial-data.sql
elif [ "$arg" = "manual" ]; then
    docker cp "VMRMEMBERS.FDB" dbtest_db_1:/firebird/data/VMRMEMBERS.FDB
fi

echo "Firebird DB is ready for operations on port 3050"

