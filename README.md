## Run
```
cp .env.example .env
edit .env
env $(cat .env | xargs) go run main.go
```
