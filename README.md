# Go api server for touchly app

## How to run

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -source file://${PWD}/scripts/migrations -database postgres://postgres:mysecretpassword@localhost:5432/touchly\?sslmode=disable up
```

```bash
cp configs/config.api.example.yaml config.yaml
go run main.go
```

```shell
kubectl create secret generic touchly-secrets --dry-run=client --from-env-file=.env -o yaml |
  kubeseal \
    --controller-name=sealed-secrets \
    --controller-namespace=kube-system \
    --format yaml >deployment/secret.yaml

```