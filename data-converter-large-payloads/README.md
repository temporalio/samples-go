This sample shows an implementation of a payload converter that automatically detects workflow and activity payloads larger than a certain threshold and writes them to a file if that's the case.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run data-converter-large-payloads/worker/main.go
```
3) Run the following command to start the example
```
go run data-converter-large-payloads/starter/main.go
```

For the inputs that exceed the Threshold, you should see them in the Temporal UI looking something like this:

```
{
  "metadata": {
    "encoding": "ZXhhbXBsZS9sb2NhbF9maWxl"
  },
  "data": "MDdkYjVhZTUtMGIyMC00OTRjLTk2NzgtNWExOGQwY2RiYmQ4"
}
```

Contents are base64 encoded.

The files specified by "data" (the decoded value) should be created locally.
