# updateServer
A simple golang https web server using a self-signed certificate

To run from a terminal on linux,

go run . -networkInterfaceName wlp0s20f3

substituting your main network interface name, in the event your machine uses a wired connection or a different Wi-Fi interface name.

On Mac OS, the server is set to use en0 as the network interface, but you can use a different network, as above, if you want to use
a different network connection.
