FROM scratch

EXPOSE 1080
ENTRYPOINT [ "/server" ]
HEALTHCHECK --retries=10 CMD /health

COPY server /
COPY health_check /bin/sh
