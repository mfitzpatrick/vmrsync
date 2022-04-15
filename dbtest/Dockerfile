FROM jacobalberty/firebird:v2.5.9-sc

ENV ISC_PASSWORD=vmrdbpass

COPY /VMRMEMBERS.FDB.sql ./setup/schema.sql
COPY /initial-data.sql ./setup/

