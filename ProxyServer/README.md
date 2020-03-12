# ProxyServer

# Proxy (Протестировано в Google Chrome)
Добавить сертификат rootCert.cert в сертификаты авторизации в браузере  
Выставить параметры прокси сервера localhost:8080  
Запустить докер контейнер следующими командами:  
git clone https://github.com/Toringol/InformationSecurity.git  
cd ProxyServer/  
sudo docker build -t proxyServerName .  
sudo docker run -p 8080:8080 -p 8090:8090 --name proxyServerName -t proxyServerName

# Repeater test
localhost:8090/history - информация о сохраненных запросах  
localhost:8090/request/{id} - запрос  

# Примеры работы находятся в папке tests