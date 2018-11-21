# fxApi


Fake API server for testing. Creates a series of HTTP endpoints for testing services that pull/scrape data.

### Install / Run

**Docker**:
```bash
docker run --rm -p 8085:8085 txn2/fxapi:latest -port=8085
```

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
