FROM docker:25.0-git

COPY mib /usr/local/bin/mib

ENTRYPOINT ["/usr/local/bin/mib"]
