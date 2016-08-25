#!/bin/bash
set -eu

output_dir=${output_dir:-}
nodes=()
extra_log_files=""

function usage(){
  >&2 echo "Usage:
  -n (Required) IP address or hostname of the mysql or proxy node to retrieve logs from (can be specified multiple times)
  -d (Required) The output directory
  -X (Optional) Include audit and binary logs

  example:
    ./download-logs -d /tmp -n 10.0.0.1 -n 10.0.0.2
  "
  exit 1
}

while getopts "d:n:X" opt; do
  case $opt in
    d)
      output_dir=$OPTARG
      ;;
    n)
      nodes+=("$OPTARG")
      ;;
    X)
      extra_log_files="-o -path '/var/vcap/store/mysql/mysql-bin.*' \
                       -o -path '/var/vcap/store/mysql_audit_logs/*'"
      ;;
    *)
      echo "Unknown arguments"
      usage
      ;;
  esac
done

if [ -z "${output_dir}" ]; then
  usage
fi

if [ ${#nodes[@]} == 0 ]; then
  usage
fi

for node in "${nodes[@]}"; do
  node_logs+="${node}.tar.gz "

  # tar exits non-zero on warnings, such as "file changed as we read it"
  # We want to still capture as much data as we can in these cases
  set +e

  ssh "vcap@${node}" "found_paths=(); \
    for path in '/var/vcap/sys/log' '/var/vcap/store/mysql' '/var/vcap/store/mysql_audit_logs'; do \
      if [ -d \${path} ]; then \
        found_paths+=(\${path}); \
      fi; \
    done; \
    find \${found_paths[@]} \( \
      -path '/var/vcap/sys/log/*' \
      -o -path '/var/vcap/store/mysql/GRA*.log' \
      ${extra_log_files} \
      \) \
      -print0 | tar --create \
      --absolute-names \
      --transform s%^%${node}% \
      --gzip \
      --null \
      --files-from -" > "${output_dir}/${node}.tar.gz"
  set -e
done

pushd "${output_dir}"
  tar -zcvf mysql-logs.tar.gz ${node_logs[@]}
  rm ${node_logs[@]}

  echo "Specify a passphrase of 6-8 words long. Do not use a private passphrase, you will need to share this passphrase with anyone who will decrypt this archive."
  gpg -c --yes --cipher-algo AES256 --symmetric --force-mdc ./mysql-logs.tar.gz
  rm mysql-logs.tar.gz
popd

echo "Encrypted logs saved at ${output_dir}/mysql-logs.tar.gz.gpg"
