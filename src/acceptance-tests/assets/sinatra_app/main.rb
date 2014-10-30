require 'json'
require 'uri'
require 'bundler'

Bundler.require

error do
  <<-ERROR
Error: #{env['sinatra.error']}

Backtrace: #{env['sinatra.error'].backtrace.join("\n")}
  ERROR
end

post '/service/mysql/:service_name/write-bulk-data' do
  client = load_mysql(params[:service_name])

  value = request.env['rack.input'].read
  megabytes = value.to_i
  one_mb_string = 'A'*1024*1024

  megabytes.times do
    begin
      # Insert 1 mb into storage_quota_testing.
      # Note that this might succeed (writing the data) but raise an error because the connection was killed
      # by the quota enforcer.
      client.query("insert into storage_quota_testing (data) values ('#{one_mb_string}');")
    rescue => e
      puts "Error trying to insert one megabyte: #{e.inspect}"
      client = load_mysql(params[:service_name])
    end
  end

  counterror = ""
  clienterror = ""

  begin
    megabytes_in_db = client.query("select count(*) from storage_quota_testing").first.values.first
  rescue Mysql2::Error => e
    puts "Error trying to count total mb in database : #{e.inspect}"
    megabytes_in_db = "unknown"
    counterror += "\nCount Error: #{e.inspect}"
  end

  max_errors = 10
  current_errors = 0
  client_closed = false
  while !client_closed && current_errors < max_errors
    begin
      client.close
      client_closed = true
    rescue Exception => e
      clienterror = "\nClient Error: #{e.inspect}"
      current_errors += 1
    end
  end

  "Database now contains #{megabytes_in_db} megabytes #{counterror} #{clienterror} \n #{current_errors}"
end

post '/service/mysql/:service_name/delete-bulk-data' do
  client = load_mysql(params[:service_name])

  value = request.env["rack.input"].read
  megabytes = value.to_i

  megabytes.times do
    begin
      client.query("delete from storage_quota_testing limit 1;")
    rescue => e
      puts "Error trying to delete one megabyte: #{e.inspect}"
      client = load_mysql(params[:service_name])
      megabytes_in_db = "#{e.inspect}"
    end
  end

  counterror = ""
  clienterror = ""

  begin
    megabytes_in_db = client.query("select count(*) from storage_quota_testing").first.values.first
  rescue Mysql2::Error => e
    puts "Error trying to count total mb in database : #{e.inspect}"
    megabytes_in_db = "unknown"
    counterror += "\nCount Error: #{e.inspect}"
  end

  max_errors = 10
  current_errors = 0
  client_closed = false
  while client_closed && current_errors < max_errors
    begin
      client.close
      client_closed = true
    rescue Exception => e
      clienterror = "\nClient Error: #{e.inspect}"
      current_errors += 1
    end
  end

  "Database now contains #{megabytes_in_db} megabytes #{counterror} #{clienterror} \n #{current_errors}"
end

get '/ping' do
  'OK'
end

get '/service/mysql/:service_name/:key' do
  client = load_mysql(params[:service_name])
  query = client.query("select data_value from data_values where id = '#{params[:key]}'")
  value = query.first['data_value'] rescue nil
  client.close
  value
end

post '/service/mysql/:service_name/:key' do
  value = request.env["rack.input"].read
  client = load_mysql(params[:service_name])

  result = client.query("select * from data_values where id = '#{params[:key]}'")
  if result.count > 0
    client.query("update data_values set data_value='#{value}' where id = '#{params[:key]}'")
  else
    client.query("insert into data_values (id, data_value) values('#{params[:key]}','#{value}');")
  end
  client.close
  value
end

# Try to open :count number of simultaneous connections to the service
get '/connections/mysql/:service_name/:count' do
  clients = []

  begin
    params[:count].to_i.times do
      clients << load_mysql(params[:service_name])
    end

    'success'
  ensure
    clients.each { |client| client.close }
  end
end

class DatabaseCredentials
  extend Forwardable
  def_delegators :@uri, :host, :port, :user, :password

  def initialize(uri)
    @uri = URI(uri)
  end

  def database_name
    @uri.path.slice(1..-1)
  end
end

def load_mysql(service_name)
  mysql_service = load_service_by_name(service_name)
  client = Mysql2::Client.new(
    :host => mysql_service['hostname'],
    :username => mysql_service['username'],
    :port => mysql_service['port'].to_i,
    :password => mysql_service['password'],
    :database => mysql_service['name']
  )
  result = client.query("SELECT table_name FROM information_schema.tables WHERE table_name = 'data_values'")
  client.query("CREATE TABLE IF NOT EXISTS data_values ( id VARCHAR(20), data_value VARCHAR(20)); ") if result.count != 1
  result = client.query("SELECT table_name FROM information_schema.tables WHERE table_name = 'storage_quota_testing'")
  client.query("CREATE TABLE IF NOT EXISTS storage_quota_testing (id MEDIUMINT NOT NULL AUTO_INCREMENT, data LONGTEXT, PRIMARY KEY (ID)); ") if result.count != 1
  client
end

def load_service_by_name(service_name)
  services = JSON.parse(ENV['VCAP_SERVICES'])
  services.values.each do |v|
    v.each do |s|
      if s["name"] == service_name
        return s["credentials"]
      end
    end
  end
  raise "service with name #{service_name} not found in bound services"
end
