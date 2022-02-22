### Steps to run this sample:

1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run worker/main.go
```
3) Run the following command to start the example
```
go run starter/main.go
```
4) Start the remote data converter, used to allow tctl to decrypt payloads
```
go run remote-data-converter/main.go
```
5) Run the following command and see that the payloads are encrypted
```
tctl workflow show --wid encryption_workflowID
```
6) Run the following command and see decrypted payloads
```
tctl --remote_data_converter_endpoint http://localhost:8081 workflow show --wid encryption_workflowID
```

### Test decryption for Temporal Web 

If you have Temporal Web running in your test environment you can try the Temporal Web support for payload decryption.
For this example we will assume that Temporal Web is running at: http://localhost:8088.
If this is not the case you should adjust the URLs below to match your setup.
1) Stop the remote data converter you started earlier if it's still running
2) Start the remote data converter, this time passing it the URL to your Temporal Web instance:
```
go run remote-data-converter/main.go http://localhost:8088
```
The Temporal Web URL is required so that the remote data converter can return CORS headers which allow the browser to talk to it.
3) Navigate to the encryption_workflowID workflow in Temporal Web and see the encrypted payloads.
4) Configure your Temporal Web session to use the remote data converter by visiting this link in your browser:
```
http://localhost:8088/remote-data-encoder/http%3A%2F%2Flocalhost%3A8081
```
5) Navigate to the encryption_workflowID workflow in Temporal Web again and now see the decrypted payloads.