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
defer func() {
	if r := recover(); r != nil {
		fmt.Println("panic forwarded to main:", r)
	}
}()

ctx := context.Background()

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
g.Wait()

fmt.Println("not going to print this because of the panic")
```

> Note: The panic redirection feature only works if you also call `g.Wait()`,
> if a panic occurs before `g.Wait()` is called it will just go through as
> it would on a normal Goroutine.

## Main Features:

This library allows you to:

1. Create several Goroutines easily and wait for them to return
2. If any of the Goroutines return an error this group cancels the context causing
   a graceful shutdown.
3. The graceful shutdown mechanism also simplifies the error handling as you can just
   wait for the `.Wait()` function to return and handle the error at that point.
4. If any of the Goroutines panics after `.Wait()` has been called the panic will
   be forwarded from the original goroutine to the waiting Goroutine causing the `.Wait()`
   function to panic.
   This is useful if you want to perform a graceful shutdown on the main goroutine for
   example.

## Helper Functions

### PeriodicWorker

The `threads.PeriodicWorker` is a useful helper function that allows you to create
a worker that will run periodically until it either returns an error or the context
is cancelled, e.g.:

```go
g := threads.NewGroup(ctx)

// A worker that runs immediately at start and then once every second:
g.Go(threads.PeriodicWorker(1*time.Second, func(ctx context.Context) error {
	fmt.Println("one second has passed: %v", time.Now())
	return nil
}))

g.Wait()
```

This worker is particularly useful because if the context is cancelled it will
perform a graceful shutdown, so you don't have to write this behavior youself.

If you want to write unit tests for this worker there is a way of mocking
the `time.After` call done inside of it:

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()

ctx = threads.ContextWithTimeAfterMock(ctx, func(triggerCh chan time.Time, waitCh chan time.Duration) {
	<-waitCh                 // Wait until time.After is called
	triggerCh <- time.Time{} // Makes <-time.After return
	<-waitCh                 // Waits again
	cancel()                 // Forces the worker to stop:
})

g := threads.NewGroup(ctx)

// A worker that runs immediately at start and
// then once again when triggerCh receives a message:
count := 0
g.Go(threads.PeriodicWorker(1*time.Hour, func(ctx context.Context) error {
	count++
	fmt.Printf("Run count: %v\n", count)
	return nil
}))

g.Wait()
```

## LICENSE

This project was created by Blackpoint Cyber to help the community, it uses
a public domain license meaning you can copy and use any part of it without
worrying about any restrictions.
