# comptest

> API will be unstable until major release.

The package aims to make component testing easier by providing a set of helper functions.

- Build binary from your code and run it
- Inject migrations with seed data
- Run migrations up and down
- Wait for database, mocks and main service to be ready
- Prepare and use gcp pubsub. Sends messages to pubsub.

## Quickstart

Get package

```bash
go get github.com/ingridhq/comptest
```

And start using it:

```go
func TestMain(t *testing.M) {
	cleanUp := comptest.MustBuildAndRun("./main.go", "./main.bin", "./comptest.logs")
	defer cleanUp()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := comptest.WaitForMainServer(ctx); err != nil {
		cleanUp()
		log.Fatalf("wait for main server failed: %v", err)
	}

	t.Run()
}

func Test_response(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code : %v", resp.StatusCode)
	}
}
```

Full examples can be found in `_example` directory. You will learn there how to use most of package functionality there. You can run all examples with `make -s` or one by one by calling `make -s` in each subdirectory.

---

## Read more

You can read more about component testing here:

- https://tech.ingrid.com/component-testing/