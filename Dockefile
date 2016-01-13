FROM golang:1.5.2

RUN go get github.com/tools/godep

ADD . /go/src/github.com/darkhelmet/ForrestFire
WORKDIR /go/src/github.com/darkhelmet/ForrestFire
RUN rm -rf Godeps/_workspace/bin Godeps/_workspace/pkg

ENV CGO_ENABLED 0

RUN godep go install ./...

ENV PATH   /go/src/github.com/darkhelmet/ForrestFire/Godeps/_workspace/bin:/go/src/github.com/darkhelmet/ForrestFire/vendor:$PATH
ENV GOPATH /go/src/github.com/darkhelmet/ForrestFire/Godeps/_workspace:$GOPATH

ENV PORT 80
EXPOSE 80

CMD /go/src/github.com/darkhelmet/ForrestFire/bin/ForrestFire
