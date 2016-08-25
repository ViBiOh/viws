FROM scratch

EXPOSE 1080
ENTRYPOINT [ "/server" ]

COPY server /
VOLUME /www
