# Generate MRQ Email Address Column
This service is a simple program which will generate MRQ email addresses
in the firebird MEMBERS table.

## Building
This needs to run on the local network, which possibly means that it needs to be
cross-compiled for running on Windows.
```
GOOS=windows GOARCH=amd64 go build -o gen_mrqemails.exe
```

## Usage
The service needs to be run with a configuration file that contains information
on how to connect to the real running DB. The environment variable `CONFIG_FILE`
should contain the path to the YAML file. To specify this in bash, run:
```
CONFIG_FILE=".config.yml" gen_mrqemails.exe
```

