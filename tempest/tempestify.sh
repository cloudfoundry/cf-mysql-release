#!/bin/sh

# This is a temporary hack to generate a tarball that Tempest will accept. If you're
# using an old version of tempest that doesn't take a tarball upload, you can use curl
# to upload the pieces instead:
#
#curl -k -u $user:$password -X POST https://$installer_ip/api/releases --form release[file]=@$release_path
#curl -k -u $user:$password -X POST https://$installer_ip/api/stemcells --form stemcell[file]=@$stemcell_path
#curl -k -u $user:$password -X POST https://$installer_ip/api/metadata --form metadata[file]=@$metadata_path
#
# TODO: figure out the real build and release process


rm -rf /tmp/tempest
mkdir -p /tmp/tempest

# bosh create release --with-tarball --force
cp dev_releases/*.tgz /tmp/tempest
cp ~/stemcells/bosh-stemcell-1116-vsphere-esxi-ubuntu.tgz /tmp/tempest
cp cf-mysql-tempest.yml /tmp/tempest

cd /tmp/tempest
tar cvpf ../cf-mysql-tempest.tar *
