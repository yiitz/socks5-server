#!bin/bash

domain="socks5overtls.org"
email="admin@socks5overtls.org"

echo "make server cert"
openssl req -new -nodes -x509 -out ./certs/server.pem -keyout ./certs/server.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=$domain/emailAddress=$email"


