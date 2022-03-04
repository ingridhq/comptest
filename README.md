# comptest

> API will be unstable until major release.

The package aims to make component testing easier by providing a set of helper functions.

- Build binary from your code and run it
- Inject migrations with seed data
- Run migrations up and down
- Wait for database, mocks and main service to be ready
- Prepare and use gcp pubsub. Send messages to pubsub.

## Quickstart

Get package

```bash
go get github.com/ingridhq/comptest
```

And start using it:

```go
func TestMain(t *testing.M) {
	// Initialize comptest lib.
	c := comptest.New(t)
	
	c.HealthChecks(
		waitfor.TCP(os.Getenv("PUBSUB_EMULATOR_HOST")),
	)
	
	c.BuildAndRun("../main.go", waitfor.HTTP(fmt.Sprintf("http://%s/readiness", cfg.ReadinessPort)))
}

func Test_response(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code : %v", resp.StatusCode)
	}
}
```

Full examples can be found in `_example` directory. There you will learn how to use most of the package's functionality. You can run all examples with `make -s` or one by one, by calling `make -s` in each subdirectory.

---

## Read more

You can read more about component testing here:

- https://tech.ingrid.com/component-testing/