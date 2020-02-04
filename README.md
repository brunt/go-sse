# go-sse
Server Sent Events in plain Go and Javascript, with a kafka container to make it extra streamy. 

[Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) are a unidirectional way to push events from a server to a client. 
The alternatives are [websockets](https://developer.mozilla.org/en-US/docs/Glossary/WebSockets) for bi-directional streaming, or have the front-end javascript poll the server for changes on a time interval.

This demo recreates a status board in a surgery center waiting room by tracking and displaying patient IDs and showing updates of where they are in the process of surgery.

### Running locally
* from top-level project folder, `docker-compose up -d` to start a kafka container
* start consumer with `go run consumer/main.go`
* start producer with `go run producer/main.go`
* watch real-time updates at `localhost:3000`

### Lessons learned
* The majority of the complexity is on the server side, where we need to keep track of how many client connections we have in order to broadcast updates to all of them.
* The format for SSE messages is `data: ${your data}\n\n` and some variations of that.
* The SSE response from the server must have the headers `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and `Connection: keep-alive`. 

### Potential improvements
* Reduce the coupling of everything.
* Fix the mime-type issue so that go can host an html file which correctly fetches a separate javascript file.
* Implement a cache that is fetched before the front-end subscribes to new events so that refreshing the page does not lose all past information.
