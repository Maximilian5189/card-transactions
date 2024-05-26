# card-transactions

```bash
docker build --tag backend .
docker run -e TOKEN=$TOKEN -p 8080:8080 backend
```

## useful commands

```bash
fly secrets set TOKEN=<val>
```

```bash
fly config env
```

```bash
fly ssh console
```
