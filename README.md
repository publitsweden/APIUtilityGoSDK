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

### APIClient
The APIClient package is a wrapper around the client and exposes easy access methods for the most commonly used API endpoints in Publit. The APICLient-struct contains methods for performing Get, Put, Post and Delete request against the Publit APIs. 

To create an API-specific implementation of the package (against the Publit APIs) is as simple as:
```Go
SpecificAPI := "someapi"
c := &APIClient.APIClient{API: SpecificAPI}
```

And to create a specific resource-package can be done as follows:
```Go
package MyResource

// Import APIClient and Endpoint
import(
    "github.com/publitsweden/APIUtilityGoSDK/APIClient"
    "github.com/publitsweden/APIUtilityGoSDK/Endpoint"
)

// Create endpoint enumeration indexes
const (
    INDEX endpoint.Endpoint = 1 + iota
    SHOW
    CREATE
    UPDATE
    DELETE
)

// Create endpoints map for "MyResource"
var endpoints = map[endpoint.Endpoint]string{
    INDEX: "api_resource",
    SHOW: "api_resource/%v",
    CREATE: "api_resource",
    UPDATE: "api_resource/%v",
    DELETE: "api_resource/%v",
}

// MyResource is a struct containing the possible endpoints for the resource
type MyResource struct {
    Endpoint endpoint.Resource
}

// New creates a new MyResource struct (pointer) and sets the endpoints
func New() *MyResource {
    r := &MyResource{
        Endpoint: endpoint.Resource{
            Endpoints: 
        }
    }

    return r
}

// A struct that represents MyResource
type MyResource struct {
    FieldA interface{} `json:field_a`
}

// Index is a method call to the index endpoint of the resource
func (r *MyResource) Index(c *APIClient.APIClient, model interface{}, queryParams ...func(q url.Values)) error {
    // Set the endpoint to INDEX
    r.Endpoint.Endpoint = INDEX
    return c.Get(r.Endpoint, model, queryParams)
}

// Show is a method call to the Show endpoint of the resource
func (r *MyResource) Show(c *APIClient.APIClient, id int, model interface{}, queryParams ...func(q url.Values)) error {
    // Set the endpoint to SHOW
    r.Endpoint.Endpoint = SHOW
    // The qualifiers are used to populate the "sprintf"-portion of the endpoint
    r.Endpoint.Qualifiers = []interface{}{id}

    // Instead of requiring a "model" to be sent to the method we could declare the method to create a "MyResource"-struct,
    // populate it and have it returned (with the return values of the method defined as (*MyResource, error)), like so:
    // model := &MyResource{}
    // err := c.Get(r.Endpoint, model, queryParams)
    // return model, error

    return c.Get(r.Endpoint, model, queryParams)
}

// Create is a method call to the create endpoint of the resource
func (r *MyResource) Create(c *APIClient.APIClient, model interface{}, result interface{}) error {
    // Set the endpoint to CREATE
    r.Endpoint.Endpoint = CREATE
    return c.Post(r.Endpoint, model, result)
}

// and so on...

```

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