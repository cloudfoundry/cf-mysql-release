function output_to_logfiles() {
  declare log_dir="$1"

  local log_basename
  log_basename="$(basename "$0")"

  exec > >(prepend_datetime >> "${log_dir}/${log_basename}.log")
  exec 2> >(prepend_datetime >> "${log_dir}/${log_basename}.err.log")
}

function prepend_datetime() {
  awk -W interactive '{ system("echo -n [$(date +\"%Y-%m-%d %H:%M:%S%z\")]"); print " " $0 }'
}
