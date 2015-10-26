## Run app
```
cp .env.example .env
edit .env
env $(cat .env | xargs) go run main.go
```

## Run tests
```
env $(cat .env | xargs) go test ./...
```
