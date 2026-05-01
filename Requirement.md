Role: Senior Developer

Context :
Telemetry Collector: consumes telemetry from the custom message queue, parses and persists it. The
implementation should support the ability to dynamically scale up/down the number of Collectors.
The Telemetry Collector actively pulls messages from the custom queue service.

each Telemetry Collector instance should use a worker pool. The system is designed with two levels of scaling: horizontal scaling by running multiple collector instances, and vertical scaling using multiple workers per instance. Workers process telemetry concurrently after messages are pulled in batches from the queue, which improves throughput, CPU utilization, and ensures backpressure is handled effectively.

Error handling (handler) & Validation (Domain layer) 

Uber fx for DI


protocol buffer
grpc 
.proto to get the message and save to db
Fetch data from custom queues 

Telemetry Collector – Agentic Specification

Design and implement a Telemetry Collector service that:

    Acts as a consumer of a custom message queue
    Receives telemetry messages via gRPC using Protocol Buffers
    Parses, validates, and persists telemetry data into PostgreSQL

Architecture: DDD + Clean Architecture

Structure the project into clear layers:

1. Domain Layer (Core)
    Defines Telemetry entity / aggregate
    Contains:
        Business rules
        Validation logic (mandatory fields, ranges, formats)
        No dependency on external libraries

Example responsibilities:

Validate telemetry payload
Enforce invariants (e.g., GPU utilization ≤ 100%)

Flow (end-to-end)
gRPC Handler (Inbound Adapter)
        ↓
Application Use Case
        ↓
Repository Interface (Port)
        ↓
Outbound Adapter (Repository Implementation)
        ↓
PostgreSQL (via :contentReference[oaicite:0]{index=0})

Layer Responsibilities
1. Inbound Adapter (Interface Layer)
    gRPC handler (generated from .proto)
        Converts:
        protobuf → application DTO / domain input
        Calls use case

👉 This is your entry point

2. Application Layer (Use Case)
    Contains orchestration logic:
        Process telemetry
        Call domain validation
        Call repository interface

Important:

Depends only on interfaces (ports), not implementations

3. Domain Layer
    Telemetry entity
    Validation rules
    No external dependencies

4. Repository Interface (Port)

Defined in application or domain layer:

type TelemetryRepository interface {
    Save(ctx context.Context, t Telemetry) error
}

5. Outbound Adapter (Infrastructure Layer)
    Implements repository interface
    Uses:
        PostgreSQL
        Bun ORM

👉 This is what you meant by adapter layer (outbound)

Where your queue fits
    Queue consumer (gRPC client/server using gRPC + Protocol Buffers)
    → acts as inbound adapter

Testing advantage

Because of this design:

    You can mock repository using GoMock
    Test use cases without DB

Validation (Domain Layer)
    📍 Where it belongs
    Domain layer only
    Not in handler, not in DB layer
🎯 Purpose
    Ensure business correctness
    Protect invariants

💡 Examples for telemetry
    Required fields (GPU ID, timestamp)
    Value ranges (utilization: 0–100)
    Format checks

🧠 How to explain
Validation lives in the domain layer because it represents business rules. This ensures invalid data never reaches the database.

Error Categories 

Define clear categories:

1. Validation Errors (❌ don’t retry)
Bad data
Handled in domain
Action: reject / dead-letter
2. Transient Errors (🔁 retry)
DB connection issues
Network failures
Action: retry with backoff
3. System Errors (⚠️ alert/log)
Unexpected failures
Action: log + monitor

Queue-specific handling (very important)

Since you’re consuming from a queue:

On error:
Error Type	Action
Validation	Drop / send to DLQ
Transient	Retry
Success	Ack message

5. Clean Flow with Error Handling
Receive message (gRPC)
        ↓
Map protobuf → domain
        ↓
Validate (Domain)
        ↓
❌ If invalid → reject / DLQ
        ↓
Save to DB
        ↓
❌ If DB error → retry
        ↓
✅ Ack message


Add a devcontainer with devcontainer file, docker file,, compose file

api/ - contain proto message telemetry, use buf gen yaml to generate proto generated code..

Message Queues is going to be a separate services.


-- Add make file (make build, run, stop, test, test-coverage)

-- K8s & Helm chart for deployment