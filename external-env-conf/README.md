# Temporal External Client Configuration Samples

This directory contains samples that demonstrate how to use the external client configuration feature.
This feature allows you to configure a client using a TOML file and/or environment variables,
decoupling connection settings from your application code.

To see a very basic use of external env config, see the [hello world sample](./../helloworld) and how it uses
`envconfig.MustLoadDefaultClientOptions`.

### Running

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run external-env-conf/worker/main.go
```
3) Run the following command to start the example
```
go run external-env-conf/starter/main.go
```

### External configuration options

There are a few options for external configuration, see the different options in
`LoadProfile()` in `externalenvconf.go`. The client config options loading is shared 
both by the worker and starter code, ensuring both clients use the same connection settings.

NOTE: Any environment variables will override any value set in a config file.

To see more details on external env configuration, including setting environment variables,
see the [environment configuration docs page](TODO).****

