This sample demonstrates how to use `errgroup` pattern to synchronise
cancellation of workflow coroutines.


This sample only uses workflow code and uses `testsuite` in the `main` function
instead of using `worker` and `starter` to simplify the code.

To run this sample, simply do:

```
go run errgroup.go
```

And output should be:

```
2022/05/10 09:44:41 DEBUG RequestCancelTimer TimerID 1
2022/05/10 09:44:41 DEBUG RequestCancelTimer TimerID 2
2022/05/10 09:44:41 DEBUG RequestCancelTimer TimerID 3
ctx error canceled
ctx error canceled
ctx error canceled
2022/05/10 09:44:41 result: expected error received: foo error
```
