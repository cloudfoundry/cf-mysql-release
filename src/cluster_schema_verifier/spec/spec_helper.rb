$LOAD_PATH << File.join(File.dirname(__FILE__), '..')
require 'bundler'
Bundler.require(:default, 'test')
require 'webmock/rspec'

Dir['lib/**/*'].each { |f| require f }

RSpec.configure do |config|

end
