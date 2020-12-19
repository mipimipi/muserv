[![PkgGoDev](https://pkg.go.dev/badge/gitlab.com/mipimipi/yuppie)](https://pkg.go.dev/gitlab.com/mipimipi/yuppie)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/mipimipi/yuppie)](https://goreportcard.com/report/gitlab.com/mipimipi/yuppie)
[![REUSE status](https://api.reuse.software/badge/gitlab.com/mipimipi/yuppie)](https://api.reuse.software/info/gitlab.com/mipimipi/yuppie)

yuppie is a library that supports building [UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play) servers in [Go](https://golang.org/)

# Installation

Run `go get -u gitlab.com/mipimipi/yuppie`.

# Usage

You can focus on the specific application logic of your server and don't have to bother with all the generic UPnP stuff. yuppie takes care of ...

* Device discovery, i.e. [SSDP](https://en.wikipedia.org/wiki/Simple_Service_Discovery_Protocol) notifications and search responses
* Provisioning of device and service descriptions
* Management of state variables
* Eventing
* Receipt and verification of service control calls, sending of responses

The yuppie server requires ...

* a device description
* service descriptions
* handler functions for HTTP and SOAP action calls

[This example](example/README.md) shows how a simple UPnP music server can be built with yuppie. You find more detailed information about how to use yuppie to build a server [here](https://pkg.go.dev/gitlab.com/mipimipi/yuppie).

## Description files

The yuppie server requires a device and service descriptions. These descriptions can come from XML files (see [2], [3] and [4] for further information; see [the example server](example/README.md) for an [ example device description](example/device.xml) and an [example description for a ContentDirectory service](example/contentdirectory.xml)). yuppie provides functions to create the input data for the server from such files.

## Configuration

Besides device and service descriptions, yuppie requires a simple configuration to create a server. If no configuration is provided the default values are used:

* All network interfaces are used by the server
* The server listens on port 8008 

## Logging

yuppie uses [logrus](https://github.com/sirupsen/logrus) for logging. It uses the logrus default configuration (i.e. output on stdout with text formatter and info level). If you don't want that, configure the output, formatter and level in your server application. This will also be adhered to by the logging of yuppie server.

# Scope and Limitations

yuppie implements the [UPnP Device Architecture version 2.0](http://www.upnp.org/specs/av/UPnP-av-ConnectionManager-v3-Service-20101231.pdf), except:

* SSDP update notifications

  yuppie does not send update notifications if network interfaces or server IP addresses changed. Thus, it's recommended to give the server a static IP address and to restart the server if network interfaces were changed.

* [Chunked transfer encoding](https://en.wikipedia.org/wiki/Chunked_transfer_encoding) that was introduced with HTTP 1.1
* Custom UPnP data types

  yuppie only supports the standard UPnP data types

* Event subscriptions for specific state variables  

  yuppie only supports event subscriptions for all evented state variables

# Further Reading

[1] [UPnP Standards and Architecture](https://openconnectivity.org/developer/specifications/upnp-resources/upnp#architectural)

[2] [UPnP Device Architecture version 2.0 (PDF)](https://openconnectivity.org/upnp-specs/UPnP-arch-DeviceArchitecture-v2.0-20200417.pdf)

[3] [UPnP ContentDirectory Service (PDF)](http://www.upnp.org/specs/av/UPnP-av-ContentDirectory-v4-Service.pdf)

[4] [UPnP ConnectionManager Service (PDF)](http://www.upnp.org/specs/av/UPnP-av-ConnectionManager-v3-Service-20101231.pdf)

[5] [UPnP MediaServer and MediaRenderer](https://openconnectivity.org/developer/specifications/upnp-resources/upnp/mediaserver4-and-mediarenderer3/)
