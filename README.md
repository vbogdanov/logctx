# Logctx

## Intro

Logctx is an utility enhancing work with the [zap](https://github.com/uber-go/zap) logger. It provides utilities to add
logging context to `context.Context` and use it when writing logs. 

```go
ctx = logctx.WithFields(ctx,
	zap.String("username", "random"),
)
// ...
logctx.From(ctx).Debug("starting operation")
```

In addition to that, the option to add the logging context to an error is provided. 
```go
func DoOperation(ctx context.Context) error {
    err := failingOp()
    if err != nil {
        // add both the logging context from `ctx` and the additional fields
        return logctx.EnhanceError(ctx, err,
			zap.String("details", "present"), 
        )
    }
}
```

```go
err := DoOperation(ctx)
if err != nil {
    logctx.ForError(ctx, err).Error("operation failed")
}
```

## Install:
```
go get github.com/vbogdanov/logctx@latest
```

## Initialization:
```go
logctx.DefaultLogger, _ = zap.NewProduction()
```

## Sugar vs No-Sugar

`logctx` supports sugared zap logger, see the API

## HTTP Middleware

Feel free to copy or take inspiration from:
```go
func AddLoggingContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := logctx.WithFields(request.Context(),
			zap.Namespace("request"),
			zap.String("Path", request.URL.Path),
		)
		request = request.WithContext(ctx)

		logctx.From(ctx).Debug("processing started")
		next.ServeHTTP(writer, request)
		logctx.From(ctx).Debug("processing complete")
	})
}
```

## golangci-lint config

If using wrapcheck linter with logctx package:

```yaml
linters-settings:
  wrapcheck:
    ignorePackageGlobs:
      # ignore this package, as it wraps errors to add logging context
      - github.com/vbogdanov/logctx

```
