#!/usr/bin/env ruby

require_relative "shared/check_travis"
require 'pp'

def last_green_build(submodule_path)
  Dir.chdir(submodule_path) do
    last_green_build_sha = TravisBuild.new([]).travis_json.detect do |b|
      b["branch"] == "master" &&
        b["result"] == 0 &&
        b["state"] == "finished" &&
        b["event_type"] == "push"
    end["commit"]

    puts "Last green sha for #{submodule_path} is #{last_green_build_sha}"

    last_green_build_sha
  end
end

def bump_submodule_code(submodule_path, sha)
  Dir.chdir(submodule_path) do
    run "git fetch origin"
    run "git checkout #{sha}"
  end

  run "git add #{submodule_path}"
end

def run(cmd)
  system(cmd) || raise("Command #{cmd} failed")
end

def uncomitted_changes?
  !system('git diff --cached --exit-code')
end

SUBMODULES = [
  "src/cf-mysql-broker",
  "src/mariadb_ctrl/src/github.com/cloudfoundry/mariadb_ctrl",
  "src/galera-healthcheck/src/github.com/cloudfoundry-incubator/galera-healthcheck",
  "src/broker-registrar"
]

SUBMODULES.each do |sub|
  last_green_broker = last_green_build(sub)
  bump_submodule_code(sub, last_green_broker)
end

run "git status"

if uncomitted_changes?
  run "./shared/staged_shortlog"
  run './shared/staged_shortlog | git commit -F -'
end
