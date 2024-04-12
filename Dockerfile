FROM docker:git

COPY mib /usr/local/mib

ENTRYPOINT ["/usr/local/mib"]
