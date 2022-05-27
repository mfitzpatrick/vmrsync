# VMR Southport database synchronisation
This software is custom-built for VMR Southport as a tool for synchronising
the local database (which is [Firebird](https://firebirdsql.org/)) with
[TripWatch](https://tripwatch-training.platformrescue.com.au).

## Getting Started
The server first must be configured with API keys for TripWatch access, as well as
the TripWatch server's location. The file structure is in YAML and its default
expected location is in `src/.config.yml`, however the file location can be changed
by setting the `CONFIG_FILE` environment variable.

Example file structure:
```
tripwatch:
  url: https://tripwatch.url.goes.here/api
  apikey: "sample API key"
```

To run a local version of the server:
```
cd src
go run .
```

## Firebird Database
As a means of testing the link to the Firebird DB, an example of the database (with
invented data) is available in the `dbtest` subdirectory. The database will be run
in a docker container. To start it and configure it with initial data:
```
bash dbtest/start.sh
```
The DB will then be accessible to port 3050 on localhost.

To stop the container, run:
```
docker-compose -f dbtest/docker-compose.yml down --rmi all
```

## Tests
This project includes test cases as examples which are automatically run in CI. To
manually run the test cases, including integration tests which operate against the
PostMan-mocked TripWatch instance, do:
```
sh ./test.sh integration
```
This helper test script will spin up versions of the Firebird DB as well as the mocked
TripWatch instance and run all the integration tests of the system against those
instances. This is a normal golang test, so at the end of the process it will either
pass or fail.

A second form of testing is pseudo-live. For this form of testing we can run a sample
copy of the Firebird DB alongside a 'live' version of TripWatch (for testing purposes
it's recommended that you use the
[training version of TripWatch](https://tripwatch-training.platformrescue.com.au)).
To start this, run:
```
sh ./test.sh manual
```
This will leave the current Firebird instance running when it exits, so that the same
instance can be tested against repeatedly (even if the app binary needs to be updated).
To stop the running Firebird instance, run:
```
sh ./test.sh clean
```

On occasion, the Firebird DB will need to be inspected to ensure that a particular update
was properly applied. A helper item has been added to this also (NB: it only works when
the DB is actually running):
```
sh ./test.sh inspect
```

