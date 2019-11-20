FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/release-engine-prototype \
    USER_UID=1001 \
    USER_NAME=release-engine-prototype

# install operator binary
COPY build/_output/bin/release-engine-prototype ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
