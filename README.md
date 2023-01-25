# Threads Helper

This helper was created to provide an easy to use tool for managing groups of Goroutines
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

It also forwards panics from the Goroutines to the
waiting Goroutine which can be useful in some situations:

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
        fmt.Println("context canceled because of the panic on the other Goroutine")
        return nil
    })

    // A panic on any of the Goroutines will cause g.Wait() to panic
    // immediately without waiting for the remaining Goroutines, but they
    // will still receive a cancel signal so they can make a graceful
    // shutdown.
    err := g.Wait()
}
```

> Note: The panic redirection feature only works if you also call `g.Wait()`,
> if a panic occurs before `g.Wait()` is called it will just go through as
> it would on a normal Goroutine.

## Features:

This library allows you to:

1. Create several Goroutines easily and wait for them to return
2. If any of the Goroutines return an error this group cancels the context causing
   a graceful shutdown.
3. If any of the Goroutines return an error the `.Wait()` method will return it
   so it is easy to handle it.
4. If any of the Goroutines panics after `.Wait()` has been called the panic will
   be forwarded from the original goroutine the the waiting Goroutine and panic again.
   This is useful if you want to perform a graceful shutdown on the main goroutine for
   example.

There is also a useful implementation of a PeriodicWorker that will repeatedly
run the input function periodically using the given interval as a period.

This worker is also useful because if the context is cancelled it will
perform a graceful shutdown, so you don't have to write this behavior youself.

## LICENSE

This project was created by Blackpoint Cyber to help the community, it uses
a public domain license meaning you can copy and use any part of it without
worrying about any restrictions.
