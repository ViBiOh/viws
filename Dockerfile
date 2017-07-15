FROM scratch

HEALTHCHECK --retries=10 CMD http://localhost:1080/health

EXPOSE 1080
ENTRYPOINT [ "/bin/sh" ]

COPY viws /bin/sh
