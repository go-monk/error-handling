Source: https://github.com/go-monk/error-handling

Go's approach to error handling is based on two ideas:

* Errors are an important part of an application's or library’s interface.
* Failure is just one of several expected behaviors.

Thus, errors are values, just like any other values returned by a function. You should therefore pay close attention to how you create and handle them.

Some functions, like [`strings.Contains`](https://pkg.go.dev/strings#Contains) or [`strconv.FormatBool`](https://pkg.go.dev/strconv#FormatBool), can never fail. If a function *can* fail, it should return an additional value. If there's only one possible cause of failure, this value is a boolean:

```go
value, ok := cache.Lookup(key)
if !ok {
    // key not found in cache
}
```

If the failure can have multiple causes, the return value is of type `error`:

```go
f, err := os.Open("/path/to/file")
if err != nil {
    return nil, err
}
```

The simplest way to create an `error` is by using `fmt.Errorf` (or `errors.New` if no formatting is needed):

```go
// (from findlinks)
if resp.StatusCode != http.StatusOK {
	return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
}
```

But since the `error` type is an interface

```go
type error interface {
    Error() string
}
```

any type that implements a method with the signature `Error() string` is considered an `error`.

## Error-handling strategies

So, how should you handle errors? Here are five strategies, roughly sorted by frequency of use.

### 1a) Propagate the error to the caller as-is

```go
// (from findlinks)
resp, err := http.Get(url)
if err != nil {
    return nil, err
}
```

### 1b) Propagate the error to the caller with additional information

```go
// (from findlinks)
doc, err := html.Parse(resp.Body)
if err != nil {
	// html.Parse is unaware of the url so we add this information
    return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
}
```

When the error eventually reaches the program's `main` function, it should present a clear causal chain, similar to this NASA accident investigation:

```
genesis: crashed: no parachute: G-switch failed: bad relay orientation
```

### 2) Retry if the error is transient

```go
// (from wait)
func WaitForServer(url string) error {
	const timeout = 1 * time.Minute
	deadline := time.Now().Add(timeout)
	for tries := 0; time.Now().Before(deadline); tries++ {
		_, err := http.Head(url)
		if err == nil {
			return nil // success
		}
		log.Printf("server not responding (%s); retrying...", err)
		time.Sleep(time.Second << uint(tries)) // exponential backoff
	}
	return fmt.Errorf("server %s failed to respond after %s", url, timeout)
}
```

### 3) Stop the program gracefully (usually from the main package)

```go
// (from wait)
if err := WaitForServer(url); err != nil {
    fmt.Fprintf(os.Stderr, "Site is down: %v\n", err)
    os.Exit(1)
}
```

or, even better:

```go
log.SetPrefix("wait: ") // command name
log.SetFlags(0)         // no timestamp

if err := WaitForServer(url); err != nil {
    log.Fatalf("Site is down: %v\n", err)
}
```

### 4) Log the error and continue (possibly with reduced functionality)

```go
if err := Ping(); err != nil {
    log.Printf("ping failed: %v; networking disabled", err)
}
```

### 5) Safely ignore the error (rare, but sometimes appropriate)

```go
dir, err := os.MkdirTemp("", "scratch")
if err != nil {
    return fmt.Errorf("failed to create temp dir: %v", err)
}

// ...use temp dir...

os.RemoveAll(dir) // ignore error; $TMPDIR is cleaned periodically
```

## Distinguishing between errors

Sometimes it's helpful to distinguish between different *kinds* of errors, not just whether an error occurred. For example, you might want to list files for which you lack permissions. The `fs` package defines several [errors](https://pkg.go.dev/io/fs#pkg-variables) that can be checked using the `errors.Is` function:

```go
// (from forbidden)
f, err := os.Open(path)
if err != nil {
	if errors.Is(err, fs.ErrPermission) {
		forbidden = append(forbidden, path)
		continue
	}
	log.Print(err)
}
```

## More

* The Go Programming Language (2016, Go 1.5)
* https://go.dev/blog/go1.13-errors
* https://go.dev/blog/errors-are-values
* https://go.dev/blog/error-handling-and-go
