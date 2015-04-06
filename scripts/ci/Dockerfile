FROM ubuntu:trusty
MAINTAINER https://github.com/cloudfoundry/cf-mysql-release

# xvfb and chromedriver are for dashboard sso tests
# nodejs, npm and libicu52 are for switchboard ui
RUN \
      apt-get update && \
      apt-get -y install --fix-missing \
            build-essential \
            curl \
            git \
            libreadline6 \
            libreadline6-dev\
            wget \
            xvfb \
            chromium-chromedriver \
            nodejs \
            npm \
            libicu52 \
      && \
      apt-get clean

# Alias nodejs as node for switchboard ui
RUN ln -s /usr/bin/nodejs /usr/bin/node

# Install ruby-install
RUN curl https://codeload.github.com/postmodern/ruby-install/tar.gz/v0.5.0 | tar xvz -C /tmp/ && \
          cd /tmp/ruby-install-0.5.0 && \
          make install

# Install Ruby
RUN ruby-install ruby 2.1.5 -- --disable-install-rdoc

# Add ruby to PATH
ENV PATH $PATH:/home/root/.gem/ruby/2.1.5/bin:/opt/rubies/ruby-2.1.5/lib/ruby/gems/2.1.5/bin:/opt/rubies/ruby-2.1.5/bin

# Install gems
RUN /opt/rubies/ruby-2.1.5/bin/gem install bosh_cli --no-rdoc --no-ri && \
          /opt/rubies/ruby-2.1.5/bin/gem install bundler --no-rdoc --no-ri

# Phantomjs is used by switchboard ui tests
RUN wget https://github.com/Pyppe/phantomjs2.0-ubuntu14.04x64/raw/master/bin/phantomjs && \
          chmod +x phantomjs && \
          mv phantomjs /usr/bin/phantomjs

# Install cf cli
RUN curl http://go-cli.s3-website-us-east-1.amazonaws.com/releases/v6.10.0/cf-linux-amd64.tgz | tar xvz -C /bin

# Install go
RUN curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar xvz -C /usr/local/

# Add golang environment variables
ENV GOPATH /root/go
ENV PATH $PATH:/usr/local/go/bin:/root/go/bin

# Create GOPATH
RUN mkdir -p $GOPATH

# Install spiff
RUN go get github.com/cloudfoundry-incubator/spiff

# Add workspace directory
RUN mkdir -p /root/workspace && ln -s /root/workspace /workspace
