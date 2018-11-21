# fxApi


Fake API server for testing. Creates a series of HTTP endpoints for testing services that pull/scrape data.

**Docker**:
```bash
docker run --rm -p 8085:8085 txn2/fxapi:latest -port=8085
```

**Endpoints**:

Description | Route | Example | Example Output
----------- | ----- | ------- | ------
Update a counter | /counter/:name/:add  | curl localhost:8085/counter/test_a/10 | 10
Number curved over one minute | /curve/:high/:std/:dec | curl localhost:8085/curve/1000/5/2 | 766.99
Epoch (unixtime) | /epoch | curl localhost:8085/epoch | 1542835090
Current second of the minute. | /second | curl localhost:8085/second | 12
String of random text | /lorem | curl localhost:8085/lorem | Pro qui tibi inveni dum qua fit donec amare illic mea regem falli contexo pro peregrinorum.
Fixed number. | /fixed-number/:num | curl localhost:8085/fixed-number/999.9 | 999.9
Random integer in range. | /random-int/:from/:to | curl localhost:8085/random-int/100/200 | 138
Prometheus Metrics | /metrics | http://localhost:8085/metrics | (Prometheus style metrics output)

### Install

MackOs (with homebrew):
```bash
brew install tap/txn2/fxapi
```

Run as command:
```bash
fxapi -port=8085
```

Get Version:
```bash
fxapi -version
```

### Development and Testing

**Run from source**

Usage:
```bash
go run ./cmd/fxapi.go -help
```

**Testing**

Start Server:
```bash
go run ./cmd/fxapi.go --port=8085
```

Get Random Int:
```bash
curl localhost:8085/random-int/10/20
```

Get two decimal curve from 0-100 with a random deviation of 5 over one minute:
```bash
curl localhost:8085/curve/100/5/2
```

Get a fixed number:
```bash
curl localhost:8085/fixed-number/50
curl localhost:8085/fixed-number/50.25
```

Get current Epoch:
```bash
curl localhost:8085/epoch
```

Get random text:
```bash
curl localhost:8085/lorem
```

Increment a counter:
```bash
curl localhost:8085/counter/test/5
curl localhost:8085/counter/test2/1
```

Get prometheus metrics:
```bash
curl localhost:8085/metrics
```

### Dependencies

Use glide to update deps:

```bash
glide update
glide install
```

### Packaging

Using `goreleaser`.

Build without releasing:
```bash
goreleaser --skip-publish --rm-dist --skip-validate
```

Release:
```bash
GITHUB_TOKEN=$GITHUB_TOKEN goreleaser --rm-dist
```
