FROM golang:1.13

EXPOSE 8080
EXPOSE 8090

COPY . .
RUN go get "github.com/Toringol/InformationSecurity/ProxyServer/certificates"
RUN go build proxy/proxy.go
RUN go build repeater/repeater.go

CMD ./proxy & ./repeater
