FROM jacobalberty/firebird:v3.0

ENV ISC_PASSWORD=vmrdbpass

COPY /VMRMEMBERS.FDB.sql ./setup/schema.sql
COPY /initial-data.sql ./setup/

