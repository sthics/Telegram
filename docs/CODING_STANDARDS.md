# Go Coding Standards

## Error Wrapping
```go
// ❌ Bad: lost context
if err != nil {
    return err
}

// ✅ Good: wrapped
if err != nil {
    return fmt.Errorf("failed to fetch user %d: %w", uid, err)
}

// Sentinel for control flow
var ErrNotFound = errors.New("not found")
if errors.Is(err, ErrNotFound) {
    return 404
}
```

## Context
Pass `ctx context.Context` as first param.  
Timeout: 2 s Redis, 5 s Postgres.

## Project Layout
Follow `github.com/golang-standards/project-layout`:
/cmd /internal /pkg /build /configs

## Graceful Shutdown
Listen to `os.Signal` → drain WS connections → wait 15 s → exit.

## Configuration
Use `github.com/kelseyhightower/envconfig` tags.  
Never commit secrets to Git.

## Dependency Injection
Accept interfaces, return concrete types.  
Keep DI graph small (no mega-container).
