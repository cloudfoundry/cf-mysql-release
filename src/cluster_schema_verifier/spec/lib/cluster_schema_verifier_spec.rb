require 'spec_helper'

describe ClusterSchemaVerifier do
  subject(:verifier) { ClusterSchemaVerifier.new(config) }

  let(:config) do
    {
      cluster_ips: ['192.0.2.1', '192.0.2.2', '192.0.2.3'],
      arbitrator_ip: '192.0.2.3',
      mysql_port: 'mysql_port',
      mysql_user: 'user',
      mysql_password: 'password',
      healthcheck_user: 'basic_user',
      healthcheck_password: 'basic_pass',
      healthcheck_port: '10000',
    }
  end
  let(:node_ip) { '192.0.2.1' }

  before do
    allow(Kernel).to receive(:sleep).with(5)
    allow(verifier).to receive(:log)
  end

  context 'when dump_sql_schema is stubbed' do
    let(:schema_string) { 'schema!' }

    before do
      allow(verifier).to receive(:dump_sql_schema).and_return(schema_string)
    end

    describe 'get_node_sha' do
      context 'when the node is healthy' do
        before do
          stub_request(:get, '192.0.2.1:10000/mysql_status').with(basic_auth: ['basic_user', 'basic_pass']).to_return(body: 'running')
        end

        it 'calls dump_sql_schema' do
          verifier.get_node_sha(node_ip)

          expect(verifier).to have_received(:dump_sql_schema).with(node_ip)
        end

        it 'returns the sha of the schema' do
          sha = verifier.get_node_sha(node_ip)
          expect(sha).to eq(Digest::SHA1.hexdigest(schema_string))
        end
      end

      context 'when the node is not healthy' do
        before do
          stub_request(:get, '192.0.2.1:10000/mysql_status').with(basic_auth: ['basic_user', 'basic_pass']).to_return(body: 'banana')
            .then.to_return(body: 'running')
            .then.to_return(body: 'stopped')
          stub_request(:any, '192.0.2.1:10000/start_mysql_single_node')
          stub_request(:any, '192.0.2.1:10000/stop_mysql')
          stub_request(:any, '192.0.2.1:10000/start_mysql_join')
        end

        it 'starts mysql in standalone mode' do
          verifier.get_node_sha(node_ip)

          expect(a_request(:post, '192.0.2.1:10000/start_mysql_single_node')
            .with(basic_auth: ['basic_user', 'basic_pass'])).to have_been_made.once
        end

        it 'waits for the running state' do
          allow(verifier).to receive(:wait_for_state)

          verifier.get_node_sha(node_ip)

          expect(verifier).to have_received(:wait_for_state).with(node_ip, 'running')
        end

        it 'leaves no trace' do
          verifier.get_node_sha(node_ip)

          expect(a_request(:post, '192.0.2.1:10000/stop_mysql')
            .with(basic_auth: ['basic_user', 'basic_pass'])).to have_been_made.once

          expect(a_request(:post, '192.0.2.1:10000/start_mysql_join')
            .with(basic_auth: ['basic_user', 'basic_pass'])).to have_been_made.once
        end
      end
    end
  end

  describe 'wait_for_state' do
    it 'queries every 5 seconds until state is reach' do
      stub_request(:get, '192.0.2.1:10000/mysql_status').with(basic_auth: ['basic_user', 'basic_pass'])
        .to_return(body: 'banana')
        .times(2)
        .then.to_return(body: 'running')

      verifier.wait_for_state(node_ip, 'running')

      expect(Kernel).to have_received(:sleep).twice

      expect(a_request(:get, '192.0.2.1:10000/mysql_status')
               .with(basic_auth: ['basic_user', 'basic_pass'])).to have_been_made.times(3)
    end

    it 'stops after 2 minutes' do
      stub_request(:get, '192.0.2.1:10000/mysql_status').with(basic_auth: ['basic_user', 'basic_pass'])
        .to_return(body: 'banana')

      expect(Kernel).to receive(:sleep).with(5).exactly(24).times

      expect{ verifier.wait_for_state(node_ip, 'running') }.to raise_error(Exception, "Timeout while starting node #{node_ip}")
    end
  end

  describe 'dump_sql_schema' do
    let(:output) { 'output' }
    let(:error) { 'error' }
    let(:status) { double('status', success?: true) }

    before do
      allow(Open3).to receive(:capture3).and_return([output, error, status])
    end

    it 'calls mysqldump with arguments' do
      verifier.dump_sql_schema('192.0.2.1')

      expect(Open3).to have_received(:capture3).with('/var/vcap/packages/mariadb/bin/mysqldump -h192.0.2.1 -Pmysql_port -uuser -ppassword --all-databases --single-transaction --no-data --compact')
    end

    it 'returns the stdout' do
      returned_schema = verifier.dump_sql_schema('192.0.2.1')

      expect(returned_schema).to eq(output)
    end

    it 'raises error when exist status is not success' do
      allow(status).to receive(:success?).and_return false

      expect { verifier.dump_sql_schema('192.0.2.1') }.to raise_error(Exception, "Error while dumping schema: #{error}")
    end

    context 'when output has auto_increment info' do
      let(:output) { 'data AUTO_INCREMENT=13 other_data'}

      it 'removes the auto_increment statements' do
        expect(verifier.dump_sql_schema('192.0.2.1')).to eq('data other_data')
      end
    end
  end

  describe 'cluster_schemas_valid?' do
    context 'when there are three mysql nodes deployed' do
      before do
        config[:arbitrator_ip] = nil
      end

      context 'and the schemas match' do
        before do
          config[:cluster_ips].each do |ip|
            allow(verifier).to receive(:get_node_sha).with(ip).and_return('fake_sha')
          end
        end

        it 'returns true' do
          expect(verifier.cluster_schemas_valid?).to be true
        end
      end

      context 'and the schemas unmatch' do
        before do
          config[:cluster_ips].each_with_index do |ip, i|
            allow(verifier).to receive(:get_node_sha).with(ip).and_return("fake_sha_#{i}")
          end
        end

        it 'returns true' do
          expect(verifier.cluster_schemas_valid?).to be false
        end
      end
    end

    context 'when there is an arbitrator' do
      context 'and the schemas match' do
        before do
          (config[:cluster_ips] - [config[:arbitrator_ip]]).each do |ip|
            allow(verifier).to receive(:get_node_sha).with(ip).and_return('fake_sha')
          end
        end

        it 'returns true' do
          expect(verifier.cluster_schemas_valid?).to be true
        end
      end

      context 'and the schemas unmatch' do
        before do
          (config[:cluster_ips] - [config[:arbitrator_ip]]).each_with_index do |ip, i|
            allow(verifier).to receive(:get_node_sha).with(ip).and_return("fake_sha_#{i}")
          end
        end

        it 'returns true' do
          expect(verifier.cluster_schemas_valid?).to be false
        end
      end
    end
  end
end
