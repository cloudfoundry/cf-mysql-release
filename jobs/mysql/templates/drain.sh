#!/bin/bash -e

<% if p('cf_mysql_enabled') == true %>
PIDFILE=/var/vcap/sys/run/mariadb_ctl/mariadb_ctl.pid
source /var/vcap/packages/cf-mysql-common/pid_utils.sh

set +e
kill_and_wait ${PIDFILE} 300 0 > /dev/null
return_code=$?

echo 0
exit ${return_code}
<% else %>
echo 0
exit 0
<% end %>
