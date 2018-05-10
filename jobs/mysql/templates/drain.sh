#!/bin/bash -e

<% if p('cf_mysql_enabled') == true %>
/var/vcap/packages/mariadb/support-files/mysql.server stop --pid-file=/var/vcap/sys/run/mysql/mysql.pid > /dev/null
return_code=$?
echo 0
exit ${return_code}
<% else %>
echo 0
exit 0
<% end %>
