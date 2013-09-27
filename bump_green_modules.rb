#!/usr/bin/env ruby

require_relative "shared/check_travis"
require 'pp'

def last_green_build(submodule_path)
  Dir.chdir(submodule_path) do
    last_green_build_sha = TravisBuild.new([]).travis_json.detect do |b|
      b["branch"] == "master" && b["result"] == 0 && b["state"] == "finished"
    end["commit"]

    puts "Last green sha for #{submodule_path} is #{last_green_build_sha}"

    last_green_build_sha
  end
end

def bump_submodule_code(submodule_path, sha)
  Dir.chdir(submodule_path) do
    run "git checkout #{sha}"
  end

  run "git add #{submodule_path}"
end

def run(cmd)
  system(cmd) || raise("Command #{cmd} failed")
end

SUBMODULE = "src/cf-mysql-broker"

last_green_broker = last_green_build(SUBMODULE)
bump_submodule_code(SUBMODULE, last_green_broker)

run "git status"
run "./shared/staged_shortlog"
run './shared/staged_shortlog | git commit -F -'
