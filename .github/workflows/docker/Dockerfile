# syntax=docker/dockerfile:1
# WARNING: For github workflow only.
# -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
FROM docker.io/ubuntu:22.04
LABEL org.opencontainers.image.authors="Nyk Ma <nykma@next.id>"
WORKDIR /app
RUN mkdir /app/config && chown -R 1000:1000 /app

USER 1000:1000
VOLUME [ "/app/config" ]
EXPOSE 9800
CMD [ "proof_server", "-config", "/app/config/config.json", "-port", "9800" ]

ARG TARGETARCH
COPY ./server_${TARGETARCH} /usr/local/bin/proof_server
