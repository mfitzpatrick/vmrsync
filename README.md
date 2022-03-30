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
live TripWatch instance, do:
```
cd src
go test --tags integration
```

