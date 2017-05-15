#!/bin/bash -e

/var/vcap/packages/mariadb/support-files/mysql.server stop > /dev/null
return_code=$?
echo 0
exit ${return_code}
