FROM vibioh/alcotest

HEALTHCHECK --retries=10 CMD http://localhost:1080/health

EXPOSE 1080
ENTRYPOINT [ "/server" ]

COPY server /
