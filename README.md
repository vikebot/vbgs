# vbgs

vbgs is the core business logic behind vikebot. It runs the game-server itself and processes game actions to render-able data for graphical interfaces

[![Build Status](https://travis-ci.org/vikebot/vbgs.svg?branch=master)](https://travis-ci.org/vikebot/vbgs)
[![codecov](https://codecov.io/gh/vikebot/vbgs/branch/master/graph/badge.svg)](https://codecov.io/gh/vikebot/vbgs)
[![Go Report Card](https://goreportcard.com/badge/github.com/vikebot/vbgs)](https://goreportcard.com/report/github.com/vikebot/vbgs)
[![GoDoc](https://godoc.org/github.com/vikebot/vbgs?status.svg)](https://godoc.org/github.com/vikebot/vbgs)

---

## Implement additional operations

### 1. Create a new OpFile

Create a new file named `opName.go` where `Name` is single identifier ([Camel-case](https://en.wikipedia.org/wiki/Camel_case)) for the operation you want to implement. For example if you want to create the operation **Radar** you create `opRadar.go` in the repositories root folder.

### 2. Packet structures

In the file created in step one now add a `namePacket` and `nameObj` struct as shown below, where `name` is the identifier of the file in [Pascal-case](https://en.wikipedia.org/wiki/PascalCase).

```go
type radarObj struct {
}

type radarPacket struct {
    Type string   `json:"type"`
    Obj  radarObj `json:"obj"`
}
```

#### Parameters to your operation

If you need to pass parameters to your operation from the client side then add them to your `nameObj` struct as pointer fields. Only if we use pointers we can differentiate between the following two cases.

1. The client didn't sent the additional parameters
2. The client sent invalid fmt/values/etc for the parameters

> Fields in the `namePacket` struct mustn't be pointers because they are checked in general by the `packetHandler` and `dispatcher`.

Below you can see an example from the move operation

```go
type moveObj struct {
    Direction *string `json:"direction"`
}
type movePacket struct {
    Type string  `json:"type"`
    Obj  moveObj `json:"obj"`
}

```

### 3. Operation endpoint

Add the function that is responsible for handling a packet of your type. This function should be named after your filename (of course without the suffix `.go`) and accept a `*ntcpclient` and a packet of your type.

```go
func opRadar(c *ntcpclient, packet radarPacket) {

}
```

### 4. Register your operation endpoint

In order to register your operation endpoint you need to add your identifier as a case in `dispatcher.go`'s packet switch. In the case you need to unmarshal the `data`, check for `err` and call your own dispatch endpoint afterwards. For good examples look at the other operations already existing.

```go
case "radar":
    var radar radarPacket
    err = json.Unmarshal(data, &radar)
    if err != nil {
        c.Respond(statusInvalidJSON)
        return
    }
    opRadar(c, radar)
    return
```

### 5. Implement your operation endpoint

You can now implement the operation itself. It should get prechecked and dispatched to your func. To get a feeling how good implementations look like look at the already existsing examples. In general you should keep a few things in mind:

- **Add testing!!!**
- Check for `nil` fields in your custom `nameObj` structs
- Don't forget to push updates to the `updatePush` network

## Underlying construction of packages

```
+--------------+-------------------+-----------------------+
| IV (16bytes) | Encrypted payload | Suffix (1byte = '\n') |
+--------------+-------------------+-----------------------+
```

### Payload

```
JsonPacket format -> Buffer -> Encrypt -> Base64
```
