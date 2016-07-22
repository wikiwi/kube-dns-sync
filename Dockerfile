FROM alpine:3.4

ARG BUILD_DATE
ARG VCS_REF
ARG VCS_VERSION

LABEL org.label-schema.build-date=${BUILD_DATE} \
      org.label-schema.vcs-ref=${VCS_REF} \
      org.label-schema.vcs-version=${VCS_VERSION} \
      org.label-schema.vcs-url="https://github.com/wikiwi/kube-dns-sync" \
      org.label-schema.vendor=wikiwi.io \
      org.label-schema.name=kube-dns-sync \
      io.wikiwi.license=MIT

COPY bin/linux/amd64/ /usr/bin/

ENTRYPOINT ["/usr/bin/kube-dns-sync"]

