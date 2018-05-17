#!/bin/bash -eu

NODE_IP=<%= spec.ip %>
MYSQL_PORT=<%= p("cf_mysql.mysql.port") %>

LOG_DIR="/var/vcap/sys/log/mysql/"

# if the node ain't running, ain't got nothin' to drain
if ! ps -p $(</var/vcap/sys/run/mysql/mysql.pid) >/dev/null; then
  echo "mysql is not running: drain OK" &>> "$LOG_DIR/drain.log"
  echo 0; exit 0 # drain success
fi

function wsrep_var() {
  local var_name=$1
  local host=$2
  local port=$3
  if [[ $var_name =~ ^wsrep_[a-z_]+$ ]]; then
    timeout 5 \
      /usr/local/bin/mysql --defaults-file=/var/vcap/jobs/mysql/config/drain.cnf -h "$host" -P "$port" \
      --execute="SHOW STATUS LIKE '$var_name'" -N |\
      awk '{print $2}' | tr -d '\n'
  fi
}

CLUSTER_NODES=(`wsrep_var wsrep_incoming_addresses $NODE_IP $MYSQL_PORT | sed -e 's/,/ /g'`)

# check if all nodes are part of the PRIMARY component; if not then
# something is terribly wrong (loss of quorum or split-brain) and doing a
# rolling restart can actually cause data loss (e.g. if a node that is out
# of sync is used to bootstrap the cluster): in this case we fail immediately.
for NODE in "${CLUSTER_NODES[@]}"; do
  NODE_IP=`echo $NODE | cut -d ":" -f 1`
  NODE_PORT=`echo $NODE | cut -d ":" -f 2`
  cluster_status=`wsrep_var wsrep_cluster_status $NODE_IP $NODE_PORT`
  if [ "$cluster_status" != "Primary" ]; then
    echo "wsrep_cluster_status of node '$NODE_IP' is '$cluster_status' (expected 'Primary'): drain failed" &>> "$LOG_DIR/drain.log"
    exit 1 # drain failed
  fi
done

# Check if all nodes are synced: if not we wait and retry
# This check must be done against *ALL* nodes, not just against the local node.
# Consider a 3 node cluster: if node1 is donor for node2 and we shut down node3
# -that is synced- then node1 is joining, node2 is donor and node3 is down: as
# a result the cluster lose quorum until node1/node2 complete the transfer!)
for NODE in "${CLUSTER_NODES[@]}"; do
  NODE_IP=`echo $NODE | cut -d ":" -f 1`
  NODE_PORT=`echo $NODE | cut -d ":" -f 2`
  state=`wsrep_var wsrep_local_state_comment $NODE_IP $NODE_PORT`
  if [ "$state" != "Synced" ]; then
    echo "wsrep_local_state_comment of node '$NODE_IP' is '$state' (expected 'Synced'): retry drain in 5 seconds" &>> "$LOG_DIR/drain.log"
    # TODO: rewrite to avoid using dynamic drain (soon to be deprecated)
    echo -5 # retry in 5 seconds
  fi
done

echo "Drain Success" &>> "$LOG_DIR/drain.log"
echo 0; exit 0 # drain success
