#! /usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# This script backs up the hwinv_by_loc table in the SMD postgres database.

set -eo pipefail

DB_NAME="hmsds"
DB_USER="postgres"
DB_TABLE="hwinv_by_loc"

BACKUP_FILE="smd_hwinv_by_loc_table_backup-$(date +"%m%d%Y-%H:%M:%S").sql"

# Determine the SMD postgres leader

echo "Determining the postgres leader..."

POSTGRES_LEADER=$(kubectl exec cray-smd-postgres-0 -n services -c postgres -t -- patronictl list -f json | jq -r '.[] | select(.Role == "Leader").Member')

echo "The SMD postgres leader is $POSTGRES_LEADER"

# Dump the contents of the hwinv_by_loc table.  We use --clean so that
# the dump file can be used to replace the table contents if necessary
# during a restore operation in the future if necessary

echo "Using pg_dump to dump the $DB_TABLE table..."

kubectl -n services exec "$POSTGRES_LEADER" -c postgres -it -- bash -c "pg_dump -U $DB_USER -d $DB_NAME -t $DB_TABLE --clean" > "$BACKUP_FILE"

echo "Dump complete. Dump file is: $BACKUP_FILE"
