# EventFlow: Distributed Event Processing System

EventFlow is a high-performance, distributed event processing system built in Go, designed to handle complex event routing with robust retry mechanisms and persistent storage capabilities.

## System Overview

### Message Flow
1. Messages enter through the Common Queue
2. Messages are distributed to specific Topic Queues
3. If Topic Queues are full:
   - Messages go to Retry Queue
   - Retry attempts are tracked
   - Exponential backoff is applied
4. After max retries or when retry queue is full:
   - Messages are moved to Secondary (Persistent) Storage
5. Failed messages can be:
   - Retried with exponential backoff
   - Declared dead after max attempts

## Architecture

```
┌─────────────┐
│   Common    │
│   Queue     │
└─────┬───────┘
      │
      ▼
┌─────────────────────┐
│    Topic Queues     │
│  ┌───┐ ┌───┐ ┌───┐ │
│  │API│ │DB │ │Mail│ │
│  └─┬─┘ └─┬─┘ └─┬─┘ │
└────┼────┼────┼────┘
     │    │    │
     ▼    ▼    ▼
   Queue Full?
     │
     ▼
┌─────────────┐
│   Retry     │
│   Queue     │
└─────┬───────┘
      │
      ▼
Retry < MaxRetries
AND Queue not full?
   │         │
   │         ▼
   │    ┌─────────────┐
   │    │ Secondary   │
   │    │ Storage     │
   │    └─────────────┘
   │
   ▼
Back to Topic Queue
with exp. backoff
```

## Features

### Queue Management
- **Common Queue**
  - Central entry point for all messages
  - Initial message validation and routing

- **Topic Queues**
  - Dedicated queues for different services (API, DB, Email)
  - Configurable queue sizes
  - Independent processing rates

### Retry Mechanism
- **Smart Retry Queue**
  - Exponential backoff strategy
  - Retry count tracking
  - Configurable maximum retries
  - Queue size monitoring

### Persistence Layer
- **Secondary Storage**
  - Persistent storage for messages that exceed retry limits
  - Storage for messages when retry queue is full
  - Message recovery capabilities
  - Audit trail for failed messages

### Monitoring
- **Metrics Collection**
  - Queue sizes and capacities
  - Retry attempts and failures
  - Processing latencies
  - Storage utilization




## Roadmap

### Phase 1 (Completed)
- [x] Basic queue implementation
- [x] Metrics collection
- [x] Basic dispatcher

### Phase 2 (Current)
- [ ] Topic queue implementation
- [ ] Retry queue with exponential backoff
- [ ] Secondary storage integration

### Future
- [ ] Web dashboard
- [ ] Dead letter queue
- [ ] Message prioritization
- [ ] Circuit breaker implementation
- [ ] Distributed deployment support

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

