 # CHAT

 a small chat app to explore golang

## Start the server
```bash
go run main.go
```


## Run some tests

Send random strings to server in parallel
```bash
 seq 10 | parallel -j 10 'head -c 500 /dev/urandom | base64 | nc -w 1 127.0.0.1 9000'
```
