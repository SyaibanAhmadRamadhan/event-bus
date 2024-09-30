# go-event-bus

`go-event-bus` is a Go library designed to provide a robust wrapper for handling publish-subscribe (pub-sub) and request-reply messaging patterns. This library features graceful shutdown and automatic reconnection mechanisms, ensuring reliable message handling even in cases where the RabbitMQ server experiences downtime. Utilizing generics, `go-event-bus` allows for dynamic interface handling, minimizing code duplication while maintaining a single defined object type.

## Features

- **Publish-Subscribe**: Easily send messages from publishers to multiple subscribers.
- **Request-Reply**: Send requests and receive responses in a structured manner.
- **Graceful Shutdown**: Ensures all ongoing operations are completed before closing connections.
- **Automatic Reconnection**: Automatically attempts to reconnect when connection errors occur.
- **Thread-safe**: Implements mutexes to secure access to shared resources.
- **Dynamic Interface Handling**: Uses Go's generics to allow dynamic message types, reducing code duplication while still depending on a single defined object type.
- **OpenTelemetry Integration**: Optionally use OpenTelemetry for observability and tracing within the messaging operations.
- **Customizable Tracing**: Enable or disable tracing as needed, giving you full control over what to trace.

## Installation


To install `go-event-bus`, you can use Go Modules by running the following command:

```bash
go get github.com/username/go-event-bus/rabbitmq@1.240930.0207
```

## Contributions 
from the community are welcome! If you have suggestions, enhancements, or bug reports, please open an issue or submit a pull request in our GitHub repository.

## Tag Versioning Example: `v1.231215.2307`
We use a time-based versioning (TBD) scheme for our releases. The format is as follows:
```txt
v1.yearMonthDate.HourMinute
```
- `year`: Last two digits of the current year (e.g., 23 for 2023).
- `month`: Two-digit month (e.g., 12 for December).
- `date`: Two-digit day of the month (e.g., 15).
- `HourMinute`: Time of release in 24-hour format, combined as HHMM (e.g., 2307 for 11:07 PM).

## Contact
For questions or support, please contact ibanrama29@gmail.com.
