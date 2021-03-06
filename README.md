# simpleSMTP
Basic implementation of a subset of the [Simple Mail Transfer Protocol (RFC 5321)](https://tools.ietf.org/html/rfc5321).

## Setting up (Linux / Mac)
*The following instructions are only valid for development / testing purpose. Do not use them for production!*
1. Generate a key: `openssl ecparam -genkey -name secp384r1 -out server.key`
2. Generate a certificate: `openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650`
3. Configure environment variables in `.env`
4. Build server: `go build`
5. Run server: `./simpleSMTP`

## Configuration
* Port: Port the server operates on
* Timeout: Amount of seconds a Session waits for a new command
* Certificate: Server's SSL certificate
* Key: Key belonging to the certificate
* Mail directory: Within this directory mails get saved. Format `<sender-hash>/<mail-hash>`
* Mail write interval: Received mails get queued and will be writen to the file systems periodically
* Maximum mail size: Maximum amount of bytes of a mail
* Maximum number of recipents: Max number of recipents per mail
* Maximum length of line: Maximum smtp protocol measage lenght
* Performance meassuring: Enables meassuring of server performance (in-server-latency)

## Testing
Test the implementation by making use of gnutls: `gnutls-cli localhost -p 443 --no-ca-verification`

## Current status of implementation

| Command       | Implemented   |           remark         |
| ------------- |:-------------:| -------------------------|
| HELO          |       ✔       |                          |
| NOOP          |       ✔       |                          |
| QUIT          |       ✔       |                          |
| MAIL          |       ✔       |                          |
| RCPT          |       ✔       |                          |
| DATA          |       ✔       |                          |
| HELO          |       ✔       |                          |
| RSET          |       ✔       |                          |
| VRFY          |       ✗       | No usermanagement so far |
| EHELO         |       ✗       | Not required so far      | 
| STARTTLS      |       ✗       | Not required so far      | 
| AUTH          |       ✗       | No usermanagement so far |
| SIZE          |       ✗       | Not required so far      | 
| HELP          |       ✗       | Not required so far      | 


## Service Level Objectives
- __inside-server-latency:__ 99% (averraged over 1 minuted) of vaild SMTP commands within an establish session on a idle system will be processed in less than 1ms

### SMTP Commands performance measurement
By setting `MEASURE_PERFORMANCE` inside `.env` to true, it is possible to measure the average response time per SMPT commands per session. After the session has been closed, the results will be printed to the CLI.

## Building Docker Image
Building a docker image with a running instance of the SMTP server is simply done by running `docker build .` within the `docker/` directory.

