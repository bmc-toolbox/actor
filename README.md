## actor


[![Go Report Card](https://goreportcard.com/badge/github.com/bmc-toolbox/actor)](https://goreportcard.com/report/github.com/bmc-toolbox/actor)


#### What?
Actor abstracts away various management actions to BMCs (Baseboard Management Controllers),
and exposes them through a consistent API.


#### Why?
When managing a large number of Baremetal servers and thier Baseboard Management Controllers,
that are from various vendors, multiple generations, we realized the management interfaces (think IPMI, RACADM etc), could use a single API to abstract away all the vendor differences.

Actions like - power on/off/cycle, pxe boot are fundamental to server asset lifecycle management,
but these are found to be unreliable, or have no return statuses, or are not exposed by each BMC in a consistent manner, nor do they behave consistently.

#### How?

Actor runs as a webservice, exposing an API, that abstracst away all
of the differences across vendors, into a single API.

Since Actor is based on bmclib, for the list of supported vendors,
https://github.com/bmc-toolbox/bmclib/blob/master/README.md

#### How to run

##### Install binary

```console
go get github.com/bmc-toolbox/actor
```

##### Docker

```console
git clone github.com/bmc-toolbox/actor
cd actor
# build docker image with application
docker-compose build actor
# start server in background, accessable by address http://localhost:8000
docker-compose up -d server
```

##### Build yourself

```console
git clone github.com/bmc-toolbox/actor
cd actor
go build
# start server, accessable by address http://localhost:8000
./actor --config actor.sample.yaml server
```

##### Check power status

A GET request on any endpoint. It is always a single action and returns the single response.

```shell
> curl -s localhost:8080/host/10.193.251.60 
{"action":"ison","status":true,"message":"ok","error":""}
```

##### Sequencing actions

Multiple actions can be chained together and sequenced. If any action fails, further actions are skipped. 

```shell
> curl -s -d '{"action-sequence": ["sleep 1s","ison"]}' localhost:8080/host/10.193.251.60 
[{"action":"sleep 1s","status":true,"message":"ok","error":""},{"action":"ison","status":true,"message":"ok","error":""}]
```

##### API return codes and responses

Code  | Info                                                          | Response
:----:|:-------------------------------------------------------------:|:------------------------------------------------------------------------------:| 
200   | All good!                                                     | `{"action":"sleep 1s","status":true,"message":"ok","error":""}`                |
400   | Request is invalid, e.g. the sequence contains unknown action | `{"error":"some error"}`                                                       |
417   | Failed to execute request.                                    | `{"action":"sleep 1s","status":false,"message":"failed","error":"some error"}` |

Single-action endpoints return one response.  
Multi-action endpoints return a list of responses but one response if the request is invalid.

##### Blade actions through Chassis BMC

This describes the Actor API endpoints to execute power related actions
on the blades through the chassis.

Endpoints

`/chassis/:host/serial/:serial`  
`/chassis/:host/position/:pos`

 
Blade actions.

Action            |  POST payload   |
:----------------:| :-------------: |
Check Powered on  | `{ "action-sequence": ["ison"] }`          |
Power On          | `{ "action-sequence": ["poweron"]}`        |
Power Off         | `{ "action-sequence": ["poweroff"]}`       |
Power Cycle       | `{ "action-sequence": ["powercycle"]}`     |
PXE Once          | `{ "action-sequence": ["pxeonce"] }`       |
Software re-seat  | `{ "action-sequence": ["reseat"] }`        |
Reset BMC         | `{ "action-sequence": ["powercyclebmc"] }` |

##### BMC actions

This describes the Actor API endpoints to execute power related
actions directly on a given BMC.

Endpoints

`/host/:host`

Action            |  POST payload   |
:----------------:| :-------------: |
Check Powered on  | `{ "action-sequence": ["ison"] }`          |
Power On          | `{ "action-sequence": ["poweron"]}`        |
Power Off         | `{ "action-sequence": ["poweroff"]}`       |
Power Cycle       | `{ "action-sequence": ["powercycle"]}`     |
PXE Once          | `{ "action-sequence": ["pxeonce"] }`       |
Reset BMC         | `{ "action-sequence": ["powercyclebmc"] }` |

`/chassis/:host`

Action            |  POST payload   |
:----------------:| :-------------: |
Check Powered on  | `{ "action-sequence": ["ison"] }`          |
Power On          | `{ "action-sequence": ["poweron"]}`        |
Power Cycle       | `{ "action-sequence": ["powercycle"]}`     |

### Build

`> go get github.com/bmc-toolbox/actor`

#### Build with vendored modules (go 1.11)

`> GO111MODULE=on go build -mod vendor -v`

#### Test

`> GO111MODULE=on go test -mod vendor ./...`

#### Lint

`> golangci-lint run ./...`

#### Notes on working with go mod

To pick a specific bmclib SHA.

`> GO111MODULE=on go get github.com/bmc-toolbox/bmclib@2d1bd1cb`

To add/update the vendor dir.

`> GO111MODULE=on go mod vendor`
