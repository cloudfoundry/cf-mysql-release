#! /bin/bash
if [ -z "$1" ]
  then
    echo "no user name supplied"
    exit 1
fi

if [ -z "$2" ]
  then
    echo "no password supplied"
    exit 1
fi

if [ -z "$3" ]
  then
    echo "no log file location supplied"
    exit 1
fi

set +e
UPGRADE_OUTPUT=$(/var/vcap/packages/mariadb/bin/mysql_upgrade -u$1 -p$2)
UPGRADE_EXIT_CODE=$?
set -e

echo "UPGRADE_OUTPUT: $UPGRADE_OUTPUT" >> $3 2>> $3
echo "UPGRADE_EXIT_CODE: $UPGRADE_EXIT_CODE" >> $3 2>> $3

echo ${UPGRADE_OUTPUT}

exit ${UPGRADE_EXIT_CODE}
