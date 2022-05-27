#! /bin/sh

BASE="$(cd $(dirname $0); pwd)"
MANUALDB=${MANUALDB:-""}

test_type=$1; shift
allowed_types="unit integration"

usage() {
    cat << EOF
Usage: $0
    $0 <unit|integration>

Tool for invoking unit or integration tests on the VMR Sync program.

EOF
}

# Check if the type is allowed
if ! echo "$allowed_types" | grep -w -q "$test_type" ; then
    usage
    exit 1
fi

if [ "$test_type" = "unit" ]; then
    cd "$BASE/src"
    go test
    exit $?
fi

if [ "$test_type" = "integration" ]; then
    cd "$BASE/src"
    CONFIG_FILE="$BASE/tripwatch-test/test-config.yml" go test -tags integration -o "$BASE/testbin" -c
    test_result=$?
    if [ "0" != "$test_result" ]; then
        echo "Go test build failure: $test_result"
        exit 1
    fi
    sh "$BASE/dbtest/start.sh"
    docker-compose -f "$BASE/tripwatch-test/docker-compose.yml" up -d
    CONFIG_FILE="$BASE/tripwatch-test/test-config.yml" "$BASE/testbin"
    test_result=$?
    docker-compose -f "$BASE/dbtest/docker-compose.yml" -f "$BASE/tripwatch-test/docker-compose.yml" logs
    if [ -n "$MANUALDB" ]; then
        docker-compose -f "$BASE/dbtest/docker-compose.yml" exec db /usr/local/firebird/bin/isql -u SYSDBA
    fi
    docker-compose -f "$BASE/dbtest/docker-compose.yml" down --rmi all
    docker-compose -f "$BASE/tripwatch-test/docker-compose.yml" down
    rm "$BASE/testbin"
    exit $test_result
fi

