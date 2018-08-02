#!/usr/bin/env bash

<% if p('cf_mysql_enabled') == true %>
set -e -o pipefail
<% if p('cf_mysql.mysql.enable_drain_healthcheck') == true %>
<%
  cluster_ips = link('mysql').instances.map(&:address)
  if_link('arbitrator') do
    cluster_ips += link('arbitrator').instances.map(&:address)
  end
%>

CLUSTER_NODES=(<%= cluster_ips.map{|e| e }.join(' ') %>)
MYSQL_PORT=<%= p("cf_mysql.mysql.port") %>
GALERA_HEALTHCHECK_PORT=<%= p("cf_mysql.mysql.galera_healthcheck.port") %>
LOG_DIR="/var/vcap/sys/log/mysql"

# If the node is not running, exit drain successfully
if ! ps -p "$(</var/vcap/sys/run/mysql/mysql.pid)" >/dev/null; then
  echo "$(date): mysql is not running: OK to drain" >> "${LOG_DIR}/drain.log"
  echo 0; exit 0 # drain success
fi

# Check the galera healthcheck endpoint on all of the nodes. If the http status returned is 000, there
# is no node at that IP, so we assume we are scaling down. If the http status returned is 200 from all nodes
# it will continue to drain. If it detects any other nodes to be unhealthy, it will fail to drain
# and exit.
for NODE in "${CLUSTER_NODES[@]}"; do
  set +e
  status_code=$(curl -s -o "/dev/null" -w "%{http_code}" "$NODE:$GALERA_HEALTHCHECK_PORT")
  set -e
  if [[ $status_code -eq 000 || $status_code -eq 200 ]]; then
    continue
  else
    echo "$(date): galera heathcheck returned $status_code; drain failed on node ${NODE}" >> "${LOG_DIR}/drain.err.log"
    exit -1
  fi
done
<% end %>
# Actually drain with a kill_and_wait on the mysql pid
PIDFILE=/var/vcap/sys/run/mariadb_ctl/mariadb_ctl.pid
source /var/vcap/packages/cf-mysql-common/pid_utils.sh

set +e
kill_and_wait "${PIDFILE}" 300 0 > /dev/null
return_code=$?

echo 0
exit ${return_code}

<% else %>
echo 0
exit 0
<% end %>
