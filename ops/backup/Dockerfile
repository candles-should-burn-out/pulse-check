FROM postgres:17-alpine

RUN apk add --no-cache bash coreutils openssh-client tar

WORKDIR /backup

COPY backup.sh /usr/local/bin/pulse-check-backup
RUN chmod +x /usr/local/bin/pulse-check-backup

ENTRYPOINT ["/usr/local/bin/pulse-check-backup"]
