#! /bin/bash

COMMAND=$1
USER=$2
PASSWORD=$3
LOG_FILE=$4

if [ -z "$1" ]
  then
    echo "no command supplied"
    exit 1
fi

if [ -z "$2" ]
  then
    echo "no user name supplied"
    exit 1
fi

if [ -z "$3" ]
  then
    echo "no password supplied"
    exit 1
fi

if [ -z "$4" ]
  then
    echo "no log file location supplied"
    exit 1
fi


set +e
COMMAND_OUTPUT=$(/var/vcap/packages/mariadb/bin/mysql -e "${COMMAND}" -u${USER} -p${PASSWORD})
COMMAND_EXIT_CODE=$?
set -e

echo "COMMAND: /var/vcap/packages/mariadb/bin/mysql -e \"${COMMAND}\" -u${USER} -p${PASSWORD}" >> ${LOG_FILE} 2>> ${LOG_FILE}
echo "COMMAND_EXIT_CODE: ${COMMAND_EXIT_CODE}" >> ${LOG_FILE} 2>> ${LOG_FILE}
echo "COMMAND_OUTPUT: ${COMMAND_OUTPUT}" >> ${LOG_FILE} 2>> ${LOG_FILE}

echo ${COMMAND_OUTPUT}

exit ${COMMAND_EXIT_CODE}
