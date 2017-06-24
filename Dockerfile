FROM scratch

EXPOSE 1080
ENTRYPOINT [ "/server" ]
HEALTHCHECK CMD /health

COPY server /
COPY health_check /health
