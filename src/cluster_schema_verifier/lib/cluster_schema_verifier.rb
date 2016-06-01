require 'open3'
require 'digest/sha1'
require 'net/http'

class ClusterSchemaVerifier
  attr_reader :cluster_ips, :arbitrator_ip, :mysql_port, :mysql_user, :mysql_password, :healthcheck_port, :healthcheck_user, :healthcheck_password

  def initialize(config)
    @cluster_ips = config[:cluster_ips]
    @mysql_port = config[:mysql_port]
    @mysql_user = config[:mysql_user]
    @arbitrator_ip = config[:arbitrator_ip]
    @mysql_password = config[:mysql_password]
    @healthcheck_port = config[:healthcheck_port]
    @healthcheck_user = config[:healthcheck_user]
    @healthcheck_password = config[:healthcheck_password]
  end

  def cluster_schemas_valid?
    (cluster_ips - [arbitrator_ip]).map do |ip|
      log "Checking node #{ip}"
      sha = get_node_sha(ip)
      log "SHA #{sha} received for node #{ip}"
      sha
    end.uniq.length == 1
  end

  def get_node_sha(node_ip)
    node_status = get_node_status(node_ip)
    log "Node Status for #{node_ip}: #{node_status}"

    broken_node = false
    if node_status != 'running'
      broken_node = true
      log "Starting node #{node_ip} in standalone mode"
      Net::HTTP.start(node_ip, healthcheck_port) do |http|
        req = Net::HTTP::Post.new('/start_mysql_single_node')
        req.basic_auth(healthcheck_user, healthcheck_password)
        http.request(req)
      end

      wait_for_state(node_ip, 'running')
    end

    sha = Digest::SHA1.hexdigest(dump_sql_schema(node_ip))

    if broken_node
      log "Stopping node #{node_ip}"
      Net::HTTP.start(node_ip, healthcheck_port) do |http|
        req = Net::HTTP::Post.new('/stop_mysql')
        req.basic_auth(healthcheck_user, healthcheck_password)
        http.request(req)
      end

      wait_for_state(node_ip, 'stopped')

      log "Starting node #{node_ip} in clustered mode"
      Net::HTTP.start(node_ip, healthcheck_port) do |http|
        req = Net::HTTP::Post.new('/start_mysql_join')
        req.basic_auth(healthcheck_user, healthcheck_password)
        http.request(req)
      end
    end

    sha
  end

  def get_node_status(node_ip)
    resp = Net::HTTP.start(node_ip, healthcheck_port) do |http|
      req = Net::HTTP::Get.new('/mysql_status')
      req.basic_auth(healthcheck_user, healthcheck_password)
      http.request(req)
    end
    resp.body
  end

  def wait_for_state(node_ip, state)
    sleep_count = 0
    node_state = get_node_status(node_ip)
    while node_state != state
      log "Waiting for #{node_ip} to be in state #{state}, current state: #{node_state}"
      sleep_count += 1
      raise "Timeout while starting node #{node_ip}" if sleep_count > 24
      Kernel.sleep(5)
      node_state = get_node_status(node_ip)
    end
  end

  def dump_sql_schema(node_ip)
    stdout, stderr, status = Open3.capture3("/var/vcap/packages/mariadb/bin/mysqldump -h#{node_ip} -P#{mysql_port} -u#{mysql_user} -p#{mysql_password} --all-databases --single-transaction --no-data --compact")

    unless status.success?
      raise "Error while dumping schema: #{stderr}"
    end

    stdout.gsub(/ AUTO_INCREMENT=\d+/, '')
  end

  def log(message)
    puts message
  end
end
