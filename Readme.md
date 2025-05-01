# EventFlow: Distributed Event Processing System

EventFlow is a high-performance, distributed event processing system built in Go, designed to handle complex event routing with robust retry mechanisms, persistent storage capabilities, and dynamic worker scaling.

## System Overview

### Message Flow
1. Messages enter through the Common Queue
2. Dispatcher routes messages to specific Topic Queues (API, DB, Email)
3. If insertion into Topic Queue fails:
   - Messages are moved to Retry Queue
   - Retry attempts are tracked with exponential backoff
   - Each retry attempts to insert into the respective channel
4. After multiple retry failures:
   - Messages are stored in persistent file storage
   - System triggers worker upscaling for the affected channel
5. Failed messages can be:
   - Retried with exponential backoff
   - Stored in file storage for later processing
   - Trigger worker scaling for improved processing

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
   │    │  File      │
   │    │ Storage    │
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
  - Real-time monitoring of queue metrics

- **Topic Queues**
  - Implemented dedicated queues for API, DB, and Email services
  - Individual monitoring systems for each queue
  - Configurable queue sizes and processing rates
  - Dynamic worker scaling based on queue load
  - Independent channel management for each service

### Retry Mechanism
- **Smart Retry Queue**
  - Exponential backoff strategy
  - Retry count tracking
  - Configurable maximum retries
  - Queue size monitoring
  - Automatic worker scaling during high retry loads
  - Persistent retry attempts with file storage backup

### Persistence Layer
- **File Storage**
  - Persistent storage for messages that exceed retry limits
  - Storage for messages when retry queue is full
  - Message recovery capabilities
  - Audit trail for failed messages
  - Automatic message archival

### Monitoring and Scaling
- **Individual Queue Monitoring**
  - Real-time metrics for each topic queue (API, DB, Email)
  - Queue sizes and capacities
  - Processing latencies
  - Error rates and retry counts
  - Worker performance metrics

- **Dynamic Worker Scaling**
  - Automatic worker scaling based on queue load
  - Configurable scaling thresholds
  - Resource utilization optimization
  - Health monitoring of worker instances
  - Channel-specific scaling triggers

## Recent Updates
- Implemented dedicated topic queues for API, DB, and Email services
- Added individual monitoring systems for each queue
- Enhanced retry mechanism with file storage backup
- Implemented automatic worker scaling based on retry failures
- Improved message persistence with file storage
- Added channel-specific monitoring and scaling

## Roadmap

###  (Completed)
- [x] Basic queue implementation
- [x] Metrics collection
- [x] Basic dispatcher
- [x] Queue monitoring
- [x] Worker scaling
- [x] Topic queue implementation
- [x] File storage integration
- [x] Enhanced retry queue with exponential backoff


### TO DO 

- [] Write tests for disatcher function
### Future
- [ ] Web dashboard
- [ ] Using time Series database like Influx Db for enhanced montitoring 
- [ ] Seperate Monitoring Service
- [ ] 
- [ ] Improved storage integration


## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

