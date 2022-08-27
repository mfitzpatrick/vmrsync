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
CONFIG_FILE="custom-config.yml" gen_mrqemails.exe
```

The default config file name is `.config.yml` and is expected to be found in the same
directory as the executable file. So if that default name is used, no environment
variable needs to be set.

### Manually Updating Email Addresses
The auto-generator is useful for a bulk-setting, but it is not perfect. Some email
addresses will not be correct, and will need to be manually set or updated for
whatever reason. This program can be used to do this via command line arguments:
```
gen_mrqemails.exe -email="newemail@mrq.org.au" -id=90701
```
NB: yes, the single `-` before the argument name is correct.

### Manually Finding Similar Emails
In order to be able to update a given email address, we must first know the user and
their associated ID. We can do some searching to find this information using this
program:
```
gen_mrqemails.exe -email="newemail%"
```
This will find all user records which have an email address containing the substring
in the email field.

