# Quickserve
- No nonsense file server
- Quickly serve static files to localhost
- Defaults to port 3000, override with --port or -p
- Serves all files/folders that are in the same directory as the binary, or at the path specified with --folder or -f

```
quickserve -p 4000 -f 'D:/temp'
```

## Security

Quickserve includes several built-in security protections:

- Path Traversal Protection: Prevents access to files outside the specified directory using `../` or encoded variations
- Method Restriction: Only allows safe HTTP methods (GET, HEAD, OPTIONS)
- Security Headers: Adds protective HTTP headers:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY` 
  - `X-XSS-Protection: 1; mode=block`
  - `Referrer-Policy: strict-origin-when-cross-origin`
- Null Byte Injection Protection: Prevents null byte attacks
- Request Logging: Logs all requests with timestamps and client IP addresses

## Run

```
go run .
```
or
```
task run
```

```
go run . -p 4000
```

## Build

```
go build -ldflags="-s -w"
```
or
```
task build
```