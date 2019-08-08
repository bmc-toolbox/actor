### actor


[![Go Report Card](https://goreportcard.com/badge/github.com/bmc-toolbox/actor)](https://goreportcard.com/report/github.com/bmc-toolbox/actor)


##### What?
Actor abstracts away various management actions to BMCs (Baseboard Management Controllers),
and exposes them through a consistent API.


#### Why?
When managing a large number of Baremetal servers and thier Baseboard Management Controllers,
that are from various vendors, multiple generations, we realized the management interfaces (think IPMI, RACADM etc), could use a single API to abstract away all the vendor differences.

Actions like - power on/off/cycle, pxe boot are fundamental to server asset lifecycle management,
but these are found to be unreliable, or have no return statuses, or are not exposed by each BMC in a consistent manner, nor do they behave consistently.

##### How?

Actor runs as a webservice, exposing an API, that abstracst away all
of the differences across vendors, into a single API.

Since Actor is based on bmclib, for the list of supported vendors,
https://github.com/bmc-toolbox/bmclib/blob/master/README.md

##### How to run

###### Install binary

```console
go get github.com/bmc-toolbox/actor
```

###### Docker

```console
git clone github.com/bmc-toolbox/actor
cd actor
# build docker image with application
docker-compose build actor
# start server in background, accessable by address http://localhost:8000
docker-compose up -d server
```

###### Build yourself

```console
git clone github.com/bmc-toolbox/actor
cd actor
go build
# start server, accessable by address http://localhost:8000
./actor --config actor.sample.yaml server
```

###### Sequencing actions.

Multiple actions can be chained together and sequenced.

for example `{ "action-sequence": ["sleep Xy", "powercycle"]}`


###### Blade actions through Chassis BMC.

This describes the Actor API endpoints to execute power related actions
on the blades through the chassis.

A GET on these endpoints will return the current power state of the blade.

Endpoints

`/chassis/:host/serial/:serial`
`/chassis/:host/position/:position`

 
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


API return codes.

Code  | Info                            |
:----:|:-------------------------------:|
200   | All good!                       |
400   | Blade isn't present in chassis. |
412   | Unable to connect to chassis.   |
417   | Failed to execute request.      |


###### BMC actions

This describes the Actor API endpoints to execute power related
actions directly on a given BMC.

A GET on this endpoint returns the current power state of the server.

Endpoints

`/host/:host`

Action            |  POST payload   |
:----------------:| :-------------: |
Check Powered on  | `{ "action-sequence": ["ison"] }`          |
Power On          | `{ "action-sequence": ["poweron"]}`        |
Power Off         | `{ "action-sequence": ["poweroff"]}`       |
Power Cycle       | `{ "action-sequence": ["powercycle"]}`     |
PXE Once          | `{ "action-sequence": ["pxeonce"] }`       |
Software re-seat  | `{ "action-sequence": ["reseat"] }`        |
Reset BMC         | `{ "action-sequence": ["powercyclebmc"] }` |

API return codes.

Code  | Info                            |
:----:|:-------------------------------:|
200   | All good!                       |
412   | Unable to connect to BMC.       |
417   | Failed to execute request.      |

#### Build

`go get github.com/bmc-toolbox/actor`

##### Build with vendored modules (go 1.11)

`GO111MODULE=on go build -mod vendor -v`

##### Notes on working with go mod

To pick a specific bmclib SHA.

`GO111MODULE=on go get github.com/bmc-toolbox/bmclib@2d1bd1cb`

To add/update the vendor dir.

`GO111MODULE=on go mod vendor`
