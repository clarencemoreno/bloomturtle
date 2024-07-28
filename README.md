# Bloomturtle: A Rate Limiter with Bloom Filter and Event-Driven Architecture

Bloomturtle is a Go-based rate limiter that leverages Bloom filters for efficient key lookups and an event-driven architecture for scalability and flexibility. It's designed to handle high-volume requests while minimizing resource consumption.

## Architecture

Bloomturtle's architecture consists of three main components:

1. **RateLimiter:** The core component responsible for enforcing rate limits. It uses a Bloom filter to track requests and a token bucket algorithm to manage the rate.
2. **EventPublisher:** A central hub for publishing events related to rate limiting, such as exceeding the rate limit or token replenishment.
3. **Storekeeper:** A listener that subscribes to events from the EventPublisher and maintains a cache of recent rate limit violations.

## Code Structure

The project is organized into the following packages:

- **internal/ratelimiter_bloom:** Contains the implementation of the Bloom filter-based rate limiter.
- **internal/storekeeper:** Implements the Storekeeper, responsible for caching rate limit violations.
- **internal/event:** Defines the event interface and provides a basic EventPublisher implementation.
- **cmd/example:** Contains example code demonstrating how to use Bloomturtle.

## Features

- **Efficient Key Lookups:** Bloom filters provide fast and space-efficient key lookups, making Bloomturtle suitable for high-volume requests.
- **Scalable and Flexible:** The event-driven architecture allows for easy integration with other systems and enables scaling by adding more listeners.
- **Customizable Rate Limits:** Bloomturtle allows you to configure the rate limit parameters, such as the capacity and rate.
- **Event-Driven Monitoring:** The EventPublisher enables you to monitor rate limiting events and build custom dashboards or alerts.

## Getting Started

### Installation

```bash
go get github.com/clarencemoreno/bloomturtle
```
Usage:
```go
import (
    "context"
    "github.com/clarencemoreno/bloomturtle/internal/ratelimiter_bloom"
)

func main() {
    // Create a new RateLimiter with a capacity of 10 tokens and a rate of 2 tokens per second.
    rl := ratelimiter_bloom.NewRateLimiter(10, 2)
    defer rl.Shutdown(context.Background())

    // Check if a request is allowed.
    if rl.Allow("key") {
        // Process the request.
    } else {
        // Handle rate limit exceeded.
    }
}
```


## Contributing
Contributions are welcome! Please open an issue or submit a pull request.

## License
Bloomturtle is licensed under the MIT License.

## Next Steps
Add more detailed documentation for each package and function.
Implement additional rate limiting algorithms.
Provide more comprehensive examples and test cases.
Explore integration with other monitoring and logging systems.