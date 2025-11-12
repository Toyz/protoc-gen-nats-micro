# Simple Python NATS Micro Example

This example demonstrates using `protoc-gen-nats-micro` to generate Python code for NATS Micro services.

## Prerequisites

- Python 3.8 or higher
- NATS Server running on `localhost:4222`
- Buf CLI installed

## Setup

1. Install Python dependencies:
```bash
cd examples/simple-py
python -m pip install -r requirements.txt
```

2. Generate the code:
```bash
# From the root of the repository
buf generate --template examples/buf-configs/buf.gen.py.yaml
```

This generates:
- `gen/example/v1/service_pb2.py` - Protobuf message definitions
- `gen/example/v1/service_nats_pb2.py` - NATS Micro server and client code
- `gen/example/v1/shared_nats_pb2.py` - Shared types (errors, interceptors, headers)

## Running

### Start the Server

```bash
cd examples/simple-py
python server.py
```

You should see:
```
Connected to NATS
ExampleService registered and running
Service name: ExampleService
Service version: 1.0.0
Service description: ExampleService

Server is running. Press Ctrl+C to stop.
```

### Run the Client

In another terminal:

```bash
cd examples/simple-py
python client.py
```

You should see the client making requests and receiving responses:
```
Connected to NATS
ExampleService client created

=== Testing Echo ===
Response: Hello from Python!
Timestamp: 1234567890
Response headers: {}

=== Testing Echo with Headers ===
Response: Hello from Python!
Response headers: {}

=== Testing GetGreeting ===
Greeting: Hello, Python Developer!
Response headers: {}

...
```

## Features Demonstrated

### Basic Service Registration
```python
# Server
handler = MyExampleService()
service = await register_example_service(nc, handler)
```

### Type-Safe Handler Implementation
```python
class MyExampleService(ExampleServiceHandler):
    async def echo(
        self,
        req: pb.EchoRequest,
        info: ServerInfo
    ) -> pb.EchoResponse:
        print(f"Headers: {info.headers}")
        return pb.EchoResponse(
            message=req.message,
            timestamp=int(time.time())
        )
```

### Client with Headers
```python
# Client
client = ExampleServiceClient(nc)
resp, headers = await client.echo(
    req,
    headers={"X-User-ID": "12345"}
)
```

### Custom Timeouts
```python
resp, headers = await client.get_greeting(
    req,
    timeout=5.0  # 5 second timeout
)
```

### Service Discovery
```python
for endpoint in client.endpoints():
    print(endpoint)
# Output:
#   example-service.echo
#   example-service.get-greeting
```

## Next Steps

For more advanced examples with interceptors and error handling, see:
- `examples/complex-server/` - Go server with interceptors
- `examples/complex-client/` - Go client with interceptors
- `examples/simple-ts/` - TypeScript implementation

For API documentation, see [API.md](../../API.md) in the repository root.
