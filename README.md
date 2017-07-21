[![GoDoc](https://godoc.org/github.com/publitsweden/APIUtilityGoSDK/common?status.svg)](https://godoc.org/github.com/publitsweden/APIUtilityGoSDK/common)

# APIUtilityGoSDK

This collection of packages includes utility functions and objects used to communicate with the Publit APIs.

## Installation
To use this within a Go project simply run:

```bash
go get github.com/publitsweden/APIUtilityGoSDK
```

## Usage
See the Godocs: 
* https://golang.org/pkg/github.com/publitsweden/APIUtilityGoSDK/APILog
* https://golang.org/pkg/github.com/publitsweden/APIUtilityGoSDK/client
* https://golang.org/pkg/github.com/publitsweden/APIUtilityGoSDK/common

for more information about implementation, examples and usage.

### Client
The client package contains the client.Client struct which aids with performing requests agains the Publit APIs.

It handles authentication and http requests.

```Go
c := client.New(
    func(c *client.Client) {
        c.User = "MyUserName"
        c.Password = "MyPassword"
    },
)
```

The client has builtin methods to authorise a request and set authorisation token per request if not previously set.

### APILog
The APILog package contains logging methods that the PublitGoSDK will use for logging internal messages.
The APILog is created automatically and bound to client.Client when creating it with client.New().

```Go
logger := APILog.New()
logger.Debug("Some debug information.")
// Outputs: 
// file.go:1: {"level":"DEBUG","message":"\"Some debug information.\"","timestamp":"2017-07-14T12:14:29.034Z"}
```

The output of the APILog is by default set to ioutil.Discard. To turn it on set the APILog.LogOutput to an io.Writer.

```Go
APILog.LogOutput = os.Stdout
```

The information of the APILog can also be modified by setting various variables:
```Go
// Sets output level by Bitmask. 
// Default is: APILog.LEVEL_DEBUG | APILog.LEVEL_INFO which logs both info and debug information.
APILog.OutputLevel = APIlog.LEVEL_DEBUG

// By default the APILog logs in json format. This can be turned off.
APILog.LogJsonFormat = false

// If additional information about the log message should be shown the same flags can be sent in to APILog as for Go's log.SetOutput()
// All of the information dictated by the log flags will be shown before the json formatted error if it has not been disabled.
APILog.LogFlags = log.Ldate | log.Ltime | log.Llongfile
```

### Common
The Common package contains helper methods and objects to use for interfacing with the PublitAPIs.

Most of the methods there will be of help when using the more specific API SDKs.