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

USER jboss