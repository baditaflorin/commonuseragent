# Common User Agent

`commonuseragent` is a Go module designed to provide an easy-to-use interface for retrieving common desktop and mobile user agents. It allows users to fetch arrays of user agents or a single random user agent from pre-defined lists.

## Features

- Retrieve a list of all desktop or mobile user agents.
- Get a random desktop or mobile user agent.
- Get any random user agent from the combined desktop and mobile lists.

## Installation

To install `commonuseragent`, you need to have Go installed on your machine. Use the following command to install this module:

```bash
go get github.com/baditaflorin/commonuseragent
```

## Usage

Below are examples of how you can use the `commonuseragent` module in your Go projects.

### Importing the Module

First, import the module in your Go file:

```go
import "github.com/baditaflorin/commonuseragent"
```

### Getting All Desktop User Agents

To retrieve all desktop user agents:

```go
desktopAgents := commonuseragent.GetAllDesktop()
```

### Getting All Mobile User Agents

To retrieve all mobile user agents:

```go
mobileAgents := commonuseragent.GetAllMobile()
```

### Getting a Random Desktop User Agent

To get a random desktop user agent:

```go
randomDesktop := commonuseragent.GetRandomDesktop()
```

### Getting a Random Mobile User Agent

To get a random mobile user agent:

```go
randomMobile := commonuseragent.GetRandomMobile()
```

### Getting Any Random User Agent

To get a random user agent from either the desktop or mobile lists:

```go
randomUserAgent := commonuseragent.GetRandomUserAgent()
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue on GitHub at [https://github.com/baditaflorin/commonuseragent](https://github.com/baditaflorin/commonuseragent).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
