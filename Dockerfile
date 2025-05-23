FROM rg.fr-par.scw.cloud/vibioh/scratch

EXPOSE 1080
EXPOSE 9090

ENV VIWS_PORT=1080

HEALTHCHECK --retries=5 CMD [ "/viws", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/viws" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

COPY mime.types /etc/mime.types
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY release/viws_${TARGETOS}_${TARGETARCH} /viws
