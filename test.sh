#! /bin/sh

BASE="$(cd $(dirname $0); pwd)"
MANUALDB=${MANUALDB:-""}

test_type=$1; shift
allowed_types="clean unit integration manual inspect mrqemail"

usage() {
    cat << EOF
Usage: $0
    $0 <clean|unit|integration|manual|inspect>

Tool for invoking unit or integration tests on the VMR Sync program.

unit: Runs unit tests with a local go compiler.
integration: Starts new docker images for the DB and TripWatch mocking services and
    runs integration tests against them with a local go compiler.
manual: Starts a docker image for the DB (if one isn't currently running) and runs
    the current binary against that DB and a live TripWatch instance.
    NB: once this has started a DB instance, the current binary can be rerun multiple
    times without the DB being stopped (so data is persisted from run to run). Stop
    the DB instance with 'clean'.
inspect: If a DB container is running, connect to the DB container with an SQL terminal
    for manually inspecting the DB state.
mrqemail: Run the program to generate MRQ email addresses for each given member found
    without an MRQ email address.
clean: Cleans up any dangling services (like a DB service started by 'manual').

EOF
}

is_db_up() {
    dbtest_id="$(docker ps -f name=dbtest_db_1 -q)"
    if [ -n "$dbtest_id" ]; then
        return 0
    fi
    return 1
}

inspect_db() {
    if is_db_up ; then
        echo "Run 'connect /firebird/data/VMRMEMBERS.FDB;' to connect to the DB. CTRL+D to exit."
        docker-compose -f "$BASE/dbtest/docker-compose.yml" exec \
            db /usr/local/firebird/bin/isql -u SYSDBA
    else
        echo "No DB instance found"
    fi
}

clean() {
    if is_db_up ; then
        docker-compose -f "$BASE/dbtest/docker-compose.yml" down --rmi all
    fi
}

# Check if the type is allowed
if ! echo "$allowed_types" | grep -w -q "$test_type" ; then
    usage
    exit 1
fi


if [ "$test_type" = "clean" ]; then
    clean
fi

if [ "$test_type" = "unit" ]; then
    cd "$BASE/src"
    go test . -config-file="$BASE/tripwatch-test/test-config.yml"
    exit $?
fi

if [ "$test_type" = "integration" ]; then
    clean  # Cleanup any previous DB instances if any can be found

    cd "$BASE/src"
    go test -tags integration -o "$BASE/testbin" -c
    test_result=$?
    if [ "0" != "$test_result" ]; then
        echo "Go test build failure: $test_result"
        exit 1
    fi
    sh "$BASE/dbtest/start.sh" "test"
    docker-compose -f "$BASE/tripwatch-test/docker-compose.yml" up -d
    "$BASE/testbin" -config-file="$BASE/tripwatch-test/test-config.yml"
    test_result=$?
    docker-compose -f "$BASE/dbtest/docker-compose.yml" -f "$BASE/tripwatch-test/docker-compose.yml" logs
    if [ -n "$MANUALDB" ]; then
        inspect_db
    fi
    docker-compose -f "$BASE/dbtest/docker-compose.yml" down --rmi all
    docker-compose -f "$BASE/tripwatch-test/docker-compose.yml" down
    rm "$BASE/testbin"
    exit $test_result
fi

if [ "$test_type" = "manual" ]; then
    if ! is_db_up ; then
        sh "$BASE/dbtest/start.sh" "test"
        user_init="$BASE/dbtest/user-init.sql"
        if [ -f "$user_init" ]; then
            docker cp "$user_init" dbtest_db_1:/setup/user-init.sql
            docker-compose -f "$BASE/dbtest/docker-compose.yml" \
                exec -T db /usr/local/firebird/bin/isql -u SYSDBA \
                -i "/setup/user-init.sql"
        fi
    fi
    cd "$BASE/src"
    go run .
fi

if [ "$test_type" = "inspect" ]; then
    inspect_db
fi

if [ "$test_type" = "mrqemail" ]; then
    cd "$BASE/newemail"
    go run . -config-file="$BASE/tripwatch-test/test-config.yml"
fi

