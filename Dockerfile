FROM alpine:latest

COPY purge-dockerhub /usr/local/bin/purge-dockerhub
USER nobody
CMD /usr/local/bin/purge-dockerhub
