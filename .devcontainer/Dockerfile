FROM quay.io/mhildenb/operator-builder:rekor-0.2

USER root

ENV PATH="$PATH:/usr/local/kubebuilder/bin:/usr/local/kubebuilder/bin:/usr/local/go/bin"
RUN echo "ho"
RUN go get -v golang.org/x/tools/gopls github.com/uudashr/gopkgs/v2/cmd/gopkgs \
    github.com/ramya-rao-a/go-outline github.com/go-delve/delve/cmd/dlv golang.org/x/lint/golint
#gopkgs go-outline dlv golint

RUN chown -R jboss /home/jboss
RUN chgrp -R root /home/jboss

ENV IMG=quay.io/mhildenb/rekor-operator:0.1

# RUN foo
USER jboss