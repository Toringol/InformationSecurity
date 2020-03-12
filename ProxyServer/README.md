# ProxyServer

# Test saving request and proxy
go run main.go
google-chrome --proxy-server=https://localhost:8080

# Repeater test
go run repeater/repeater.go
localhost:8090/history - информация о сохраненных запросах
localhost:8090/request/{id} - запрос