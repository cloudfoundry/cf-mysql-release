#!/usr/bin/env bash

set -e -o pipefail

<%
  require "shellwords"

  cluster_ips = link('mysql').instances.map(&:address)
  if_link('arbitrator') do
    cluster_ips += link('arbitrator').instances.map(&:address)
  end
%>

CLUSTER_NODES=(<%= cluster_ips.map{|e| Shellwords.escape e}.join(' ') %>)
MYSQL_PORT=<%= Shellwords.escape p("cf_mysql.mysql.port") %>

function prepend_datetime() {
  awk -W interactive '{ system("echo -n [$(date +%FT%T%z)]"); print " " $0 }'
}

function wsrep_var() {
  local var_name="$1"
  local host="$2"
  if [[ $var_name =~ ^wsrep_[a-z_]+$ ]]; then
    timeout 5 \
      /usr/local/bin/mysql --defaults-file=/var/vcap/jobs/mysql/config/drain.cnf -h "$host" -P "$MYSQL_PORT" \
      --execute="SHOW STATUS LIKE '$var_name'" -N \
      | awk '{print $2}' \
      | tr -d '\n'
  fi
}

LOG_DIR="/var/vcap/sys/log/mysql"

exec 3>&1
exec \
  1> >(prepend_datetime >> $LOG_DIR/drain.out.log) \
  2> >(prepend_datetime >> $LOG_DIR/drain.err.log)

# if the node ain't running, ain't got nothin' to drain
if ! ps -p $(</var/vcap/sys/run/mysql/mysql.pid) >/dev/null; then
  echo "mysql is not running: drain OK"
  echo 0 >&3; exit 0 # drain success
fi

# Check each cluster node's availability.
# Jump to next node if unreachable(timeout 5 sec), then do not add it as test component.
# Node may have been deleted or mysql port has been updated.
for NODE in "${CLUSTER_NODES[@]}"; do
  { nc -zv -w 5 $NODE $MYSQL_PORT \
  && CLUSTER_TEST_NODES=(${CLUSTER_TEST_NODES[@]} $NODE); } \
  || continue
done

# Check if all nodes are part of the PRIMARY component; if not then
# something is terribly wrong (loss of quorum or split-brain) and doing a
# rolling restart can actually cause data loss (e.g. if a node that is out
# of sync is used to bootstrap the cluster): in this case we fail immediately.
for TEST_NODE in "${CLUSTER_TEST_NODES[@]}"; do
  cluster_status=$(wsrep_var wsrep_cluster_status "$TEST_NODE")
  if [ "$cluster_status" != Primary ]; then
    echo "wsrep_cluster_status of node '$TEST_NODE' is '$cluster_status' (expected 'Primary'): drain failed"
    exit -1 # drain failed
  fi
done

# Check if all nodes are synced: if not we wait and retry
# This check must be done against *ALL* nodes, not just against the local node.
# Consider a 3 node cluster: if node1 is donor for node2 and we shut down node3
# -that is synced- then node1 is joining, node2 is donor and node3 is down: as
# a result the cluster lose quorum until node1/node2 complete the transfer!)
for TEST_NODE in "${CLUSTER_TEST_NODES[@]}"; do
  state=$(wsrep_var wsrep_local_state_comment "$TEST_NODE")
  if [ "$state" != Synced ]; then
    echo "wsrep_local_state_comment of node '$TEST_NODE' is '$state' (expected 'Synced'): retry drain in 5 seconds"
    # TODO: rewrite to avoid using dynamic drain (soon to be deprecated)
    echo -5 >&3; exit 0 # retry in 5 seconds
  fi
done

echo "Drain Success"
echo 0 >&3; exit 0 # drain success
