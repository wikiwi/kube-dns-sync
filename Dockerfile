FROM alpine:3.4

COPY bin/linux/amd64/ /usr/bin/

ENTRYPOINT ["/usr/bin/kube-dns-sync"]
