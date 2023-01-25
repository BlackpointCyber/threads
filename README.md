# Threads Helper

This helper was created to provide an easy to use tool for managing groups of goroutines
in common use-cases.

Example usage:

```go
g := threads.NewGroup(ctx)

g.Go(func(ctx context.Context) error {
	return DoSomeTask(ctx)
})

g.Go(threads.PeriodicWorker(10*time.Second, func(ctx context.Context) error {
	fmt.Println("every 10 seconds this message will show up")
	return nil
}))

err := g.Wait()
```

It also forwards panics from the goroutines to the
waiting goroutine which can be useful in some situations:

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("panic forwarded to main:", r)
        }
    }
    g := threads.NewGroup(ctx)

    g.Go(func(ctx context.Context) error {
	    panic("foo")
    })

    g.Go(func(ctx context.Context) error {
        <-ctx.Done()
        fmt.Println("context canceled because of the panic on the other goroutine")
        return nil
    })

    // A panic on any of the Goroutines will cause g.Wait() to panic
    // immediately without waiting for the remaining Goroutines, but they
    // will still receive a cancel signal so they can make a graceful
    // shutdown.
    err := g.Wait()
}
```

## Features:

This library allows you to:

1. Create several goroutines easily and wait for them to return
2. If any of the goroutines return an error this group cancels the context causing
   a graceful shutdown.
3. If any of the goroutines return an error the `.Wait()` method will return it
   so it is easy to handle it.

There is also a useful implementation of a PeriodicWorker that will repeatedly
run the input function periodically using the given interval as a period.

This worker is also useful because if the context is cancelled it will
perform a graceful shutdown, so you don't have to write this behavior youself.

## LICENSE

This project was created by Blackpoint Cyber to help the community, it uses
a public domain license meaning you can copy and use any part of it without
worrying about any restrictions.
