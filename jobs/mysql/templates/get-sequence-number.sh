#!/usr/bin/env bash

source /var/vcap/packages/cf-mysql-common/pid_utils.sh

set -eu

read -p "This script stops mysql. Are you sure? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi

pid=""
if [ -f "/var/vcap/sys/run/mysql/mysql.pid" ]; then
  pid=$(head -1 "/var/vcap/sys/run/mysql/mysql.pid")
fi

monit stop mariadb_ctrl

if [ -n "${pid}" ]; then
  while pid_is_running "${pid}"; do
    sleep 1
  done
fi

regex="[0-9]+$"

seq_no=$(cat /var/vcap/store/mysql/grastate.dat | grep 'seqno:')

if [ "$seq_no" = "seqno:   -1" ]; then
   /var/vcap/packages/mariadb/bin/mysqld --wsrep-recover
   seq_no=$(grep "Recovered position" /var/vcap/sys/log/mysql/mysql.err.log | tail -1)
fi

if [[ "$seq_no" =~ $regex ]]; then
  instance_id=$(cat /var/vcap/instance/id)
  json_output="{\"sequence_number\":${BASH_REMATCH},\"instance_id\":\"${instance_id}\"}";
fi


echo $json_output

