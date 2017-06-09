FROM golang:alpine

ADD . /go/src/acronis-gsuite-backup
ADD ./config.json /srv
ADD ./Acronis-backup-project-8b80e5be7c37.json /srv
ADD ./google7ded6bed08ed3c1b.html /srv
ADD ./privkey.pem /srv
ADD ./cert.pem /srv


RUN \
    apk add --no-cache bash && \
    apk add --no-cache git && \
    cd /go/src/acronis-gsuite-backup/main && \
    go get -v && \
    apk del git && \
    go build -o /srv/acronis-gsuite-backup && \
    rm -rf /go/src/*

EXPOSE 1443

WORKDIR /srv

CMD ["/srv/acronis-gsuite-backup"]