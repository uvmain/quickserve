# Quickserve

- Quickly serve static files to localhost.
- Defaults to port 3000, override with --port or -p
- Serves all files/folders that are in the same directory as the binary, or at the path specified with --folder or -f

```
quickserve -p 4000 -f 'D:/source/temp'
```

## Run

```
go run .
```

```
go run . -p 4000
```

## Build

```
go build -ldflags="-s -w"
```