FROM golang:1.12 as builder

WORKDIR /app
COPY . .

RUN make viws \
 && curl -s -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem

ARG CODECOV_TOKEN
RUN curl -s https://codecov.io/bash | bash

FROM scratch

EXPOSE 1080

HEALTHCHECK --retries=10 CMD [ "/viws", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/viws" ]

ARG APP_VERSION
ENV VERSION=${APP_VERSION}

COPY --from=builder /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/bin/viws /viws
