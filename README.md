# simpleSMTP
Basic implementation of a subset of the [Simple Mail Transfer Protocol (RFC 5321)](https://tools.ietf.org/html/rfc5321).

## Setting up (Linux / Mac)
*The following instructions are only valid for development / testing purpose. Do not use them for production!*
1. Generate a key: `openssl ecparam -genkey -name secp384r1 -out server.key`
2. Generate a certificate: `openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650`
3. Configure environment variables in `.env`
4. Build server: `go build`
5. Run server: `./simpleSMTP`

## Testing
Test the implementation by making use of gnutls: `gnutls-cli localhost -p 443 --no-ca-verification`
