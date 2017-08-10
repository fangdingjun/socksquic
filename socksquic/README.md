socksquic
========


generate self-signed certificate or use exists certificate

    openssl genrsa -out private.pem 2048
    openssl req -new -x509 -key private.pem -out server.pem


run server

    ./socksquic -cert server.pem -keyfile private.pem -port 4321
