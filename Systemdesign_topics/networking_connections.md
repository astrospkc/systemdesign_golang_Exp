## Your Go API makes 1,000 outbound requests per second. The downstream service responds in 30 ms, but your API experiences 200 ms latency. 

- Outbound request rate: 1,000 requests/second
- Downstream response time: 30 ms
- Your API latency: 200 ms
- Unexpected additional latency: approximately 170 ms

#### What could cause the latency?

##### Possibilities:
- DNS Resltion : Time spent resolving the downstream hostname
- TCP connection establishment : 	New TCP connections require handshakes
- TLS handshake : New HTTPS connections require TLS negotiation
- Connection pool exhaustion :	Requests wait for an available connection
- Downstream processing :The service may be slower under concurrent load
- Response body handling : 	Inefficient reading or failure to close bodies can prevent connection reuse
- Network delays :	Packet loss, retransmissions, routing or congestion

