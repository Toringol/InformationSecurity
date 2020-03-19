FROM golang:1.13

LABEL maintainer="Shelepev Sergey"

COPY . .

RUN cd scanner && go build

CMD ./scanner/scanner