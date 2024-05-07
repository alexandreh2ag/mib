FROM docker:git

COPY mib /usr/local/bin/mib

ENTRYPOINT ["/usr/local/bin/mib"]
