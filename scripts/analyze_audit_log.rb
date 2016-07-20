#!/usr/bin/env ruby

############### Sample commands: ####################
# To see all spaces for a given app name:
#   jq '[.[] | select(.app_name == "APPNAME")] | group_by(.space) | .[] | {space: .[0].space, count: length}' < OUTPUT_FILE
#
# To see all orgs for a given app name:
#   jq '[.[] | select(.app_name == "tracker-app")] | group_by(.org) | .[] | {org: .[0].org, count: length}' < OUTPUT_FILE
#
# To see all apps of a given name and the number of entries related to that app:
#   jq 'group_by(.app_name) | map({app_name .[0].app_name, count: length})' < OUTPUT_FILE
######################################################

require 'multi_json'
MultiJson.engine

require 'active_support/all'

audit_log_file = ARGV[0]

if audit_log_file.nil?
  puts "USAGE: \"analyze_audit_log.rg PATH_TO_AUDIT_LOG_FILE > OUTPUT_FILE 2> ERR_OUTPUT\""
  exit
end

class Fetcher
  cattr_reader :query_count
  @@query_count = 0

  def self.fetch(url)
    @@query_count += 1
    MultiJson.load(`cf curl #{url}`)
  end
end

class Streamer
  def stream(str)
    unless @stream_started
      puts '['
      @stream_started = true
    end
    if @previous_stream
      puts "#{@previous_stream},"
    end

    @previous_stream = str
  end

  def close
    puts @previous_stream
    puts ']'
  end
end

class AuditEntry
  attr_accessor :timestamp,:serverhost,:username,:host,:connectionid,:queryid,:operation,:database,:object,:retcode

  def initialize(row, index)
    @timestamp,@serverhost,@username,@host,@connectionid,@queryid,@operation,@database,@object,@retcode = row
    @index = index
  end

  def to_json
    if user_provisioned_db?
      MultiJson.dump({
        line_number: @index,
        timestamp: timestamp,
        database: database,
        query: object,
        app_or_key_name: app_or_key_name,
        space: space_name,
        org: org_name,
      })
    else
      MultiJson.dump({
        line_number: @index,
        timestamp: timestamp,
        database: database,
        query: object,
        app_or_key_name: 'n/a',
        space: 'n/a',
        org: 'n/a',
      })
    end
  end

  def app_or_key_name
    si = service_instance
    if si.exists?
      si.app_or_key_for_username(username).name
    else
      si.description
    end
  end

  def space_name
    service_instance.exists? ? space.name : 'SPACE NOT FOUND'
  end

  def org_name
    service_instance.exists? ? org.name : 'ORG NOT FOUND'
  end

  def user_provisioned_db?
    database[0..2] == 'cf_'
  end

  private

  def service_instance
    ServiceInstance.for_guid(service_instance_guid)
  end

  def space
    Space.for_guid(service_instance.space_guid)
  end

  def org
    Org.for_guid(space.org_guid)
  end

  def service_instance_guid
    return nil if database.blank?
    database[3..-1].gsub('_', '-')
  end
end

class NilServiceInstance
  def exists?
    false
  end

  def description
    nil
  end
end

class ServiceInstance
  @@service_instances = {}
  def self.for_guid(guid)
    return NilServiceInstance.new unless guid
    @@service_instances[guid] ||= new(guid)
  end

  def initialize(guid)
    @json = Fetcher.fetch("/v2/service_instances/#{guid}")
  end

  def space_guid
    @json['entity']['space_guid']
  end

  def app_or_key_for_username(username)
    service_bindings.each do |binding|
      return binding.app if binding.username == username
    end
    service_keys.each do |key|
      return key if key.username == username
    end
    return NilApp.new
  end

  def exists?
    @json['error_code'].nil?
  end

  def description
    @json['description'] if !exists?
  end

  private

  def service_bindings
    return @service_bindings if @service_bindings
    bindings_json = Fetcher.fetch(@json['entity']['service_bindings_url'])
    @service_bindings = bindings_json['resources'].map do |binding|
      ServiceBinding.new(binding)
    end
  end

  def service_keys
    return @service_keys if @service_keys
    keys_json = Fetcher.fetch(@json['entity']['service_keys_url'])
    @service_keys = keys_json['resources'].map do |key|
      ServiceKey.new(key)
    end
  end
end

class ServiceBinding
  def initialize(json)
    @json = json
  end

  def username
    @json['entity']['credentials']['username']
  end

  def app
    App.for_guid(@json['entity']['app_guid'])
  end
end

class ServiceKey
  def initialize(json)
    @json = json
  end

  def username
    @json['entity']['credentials']['username']
  end

  def name
    @json['entity']['name']
  end
end

class NilApp
  def name
    'APP NOT FOUND'
  end
end

class App
  @@apps = {}
  def self.for_guid(guid)
    @@apps[guid] ||= new(guid)
  end

  def initialize(guid)
    @json = Fetcher.fetch("/v2/apps/#{guid}")
  end

  def name
    @json['entity']['name']
  end
end

class Space
  @@spaces = {}
  def self.for_guid(guid)
    @@spaces[guid] ||= new(guid)
  end

  def initialize(guid)
    @json = Fetcher.fetch("/v2/spaces/#{guid}")
  end

  def org_guid
    @json['entity']['organization_guid']
  end

  def name
    @json['entity']['name']
  end
end

class Org
  @@orgs = {}
  def self.for_guid(guid)
    @@orgs[guid] ||= new(guid)
  end

  def initialize(guid)
    @json = Fetcher.fetch("/v2/organizations/#{guid}")
  end

  def name
    @json['entity']['name']
  end
end

STDERR.puts 'Starting to read file. Making API calls. This may take some time...'

line_number = 0
streamer = Streamer.new

File.foreach(audit_log_file) do |row|
  all_columns = row.split(',')
  columns = all_columns[0..7]
  columns.push all_columns[8..-2].join(',')
  columns.push all_columns[-1]
  line_number += 1
  STDERR.puts "#{line_number}" if line_number % 1000 == 0
  streamer.stream AuditEntry.new(columns, line_number).to_json
end

streamer.close

STDERR.puts "Queries run: #{Fetcher.query_count}"
