FROM quay.io/mhildenb/operator-builder:rekor-0.2

USER root

ENV PATH="$PATH:/usr/local/kubebuilder/bin:/usr/local/kubebuilder/bin:/usr/local/go/bin"
ENV GOPATH=/home/jboss/.che
RUN go get -v golang.org/x/tools/gopls github.com/uudashr/gopkgs/v2/cmd/gopkgs \
    github.com/ramya-rao-a/go-outline github.com/go-delve/delve/cmd/dlv golang.org/x/lint/golint

# RUN mkdir /home/jboss/.che
RUN chown -R jboss /home/jboss
RUN chgrp -R root /home/jboss

# ENV IMG=quay.io/mhildenb/rekor-operator:0.1

# install a suitably old version yq (version 4.x and above won't work)
# See info here: https://mikefarah.gitbook.io/yq/upgrading-from-v3
# if you want to upgrade to the new 4.x syntax
RUN wget https://github.com/mikefarah/yq/releases/download/3.4.1/yq_linux_386 -O /usr/bin/yq &&\
    chmod +x /usr/bin/yq
RUN wget https://github.com/go-swagger/go-swagger/releases/download/v0.26.1/swagger_linux_386 -O /usr/bin/swagger &&\
    chmod +x /usr/bin/swagger

# NOTE: You might need to use subscription manager first for this to work
RUN yum install -y pinentry

RUN curl "https://d1vvhvl2y92vvt.cloudfront.net/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && ./aws/install && rm awscliv2.zip && rm -rf ./aws

USER jboss

COPY .zshrc.example $HOME/.zshrc
ENV DEMO_HOME /workspaces/rekor