# SiteScanner
Trying to find hidden get parameters from site

# How to use
git clone https://github.com/Toringol/InformationSecurity.git  
cd SiteScanner/  
sudo docker build -t scanner .  
sudo docker run -i -t --name scanner -t scanner

# How it works
It takes params from file and trying to make get request
If request has status 200 and has different content length
It will be added to check list