FROM golang:1.13

LABEL maintainer="Shelepev Sergey"

COPY . .

RUN go get "github.com/gorilla/mux"
RUN go get "github.com/Toringol/InformationSecurity/ProxyServer/certificates"

RUN cd proxy && go build
RUN cd repeater && go build

EXPOSE 8080
EXPOSE 8090

CMD ./proxy/proxy & ./repeater/repeater
