FROM golang:latest

ADD . /project
WORKDIR /project
RUN go build .
RUN mv ./dns-yml /dns-yml
RUN rm -rf /project

ENTRYPOINT ["/dns-yml"]
