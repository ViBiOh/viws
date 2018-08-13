FROM golang:1.10 as builder

ENV APP_NAME viws
ENV WORKDIR ${GOPATH}/src/github.com/ViBiOh/viws

WORKDIR ${WORKDIR}
COPY ./ ${WORKDIR}/

RUN make ${APP_NAME} \
 && mkdir -p /app \
 && curl -s -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem \
 && cp bin/${APP_NAME} /app

FROM scratch

HEALTHCHECK --retries=10 CMD [ "/viws", "-url", "https://localhost:1080/health" ]

EXPOSE 1080
ENTRYPOINT [ "/viws" ]

COPY --from=builder /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/viws /viws
