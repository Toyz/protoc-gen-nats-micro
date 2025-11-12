"""
Simple NATS Micro server example using protoc-gen-nats-micro
"""

import asyncio
import time
import sys
from pathlib import Path

# Add generated code to path
sys.path.insert(0, str(Path(__file__).parent / "gen"))

import nats
from example.v1.service_nats_pb2 import (
    ExampleServiceHandler,
    register_example_service,
)
from example.v1 import service_pb2 as pb
from example.v1.shared_nats_pb2 import ServerInfo


class MyExampleService(ExampleServiceHandler):
    """Implementation of ExampleService"""
    
    async def echo(
        self,
        req: pb.EchoRequest,
        info: ServerInfo
    ) -> pb.EchoResponse:
        """Echo the message back with a timestamp"""
        print(f"Echo called with message: {req.message}")
        print(f"Request headers: {info.headers}")
        
        return pb.EchoResponse(
            message=req.message,
            timestamp=int(time.time())
        )
    
    async def get_greeting(
        self,
        req: pb.GetGreetingRequest,
        info: ServerInfo
    ) -> pb.GetGreetingResponse:
        """Return a personalized greeting"""
        print(f"GetGreeting called for: {req.name}")
        print(f"Request headers: {info.headers}")
        
        greeting = f"Hello, {req.name}!"
        return pb.GetGreetingResponse(greeting=greeting)


async def main():
    """Main server function"""
    # Connect to NATS
    nc = await nats.connect("nats://localhost:4222")
    print("Connected to NATS")
    
    # Create service handler
    handler = MyExampleService()
    
    # Register the service
    service = await register_example_service(nc, handler)
    print(f"ExampleService registered and running")
    
    # Get service info
    info = service.info()
    print(f"Service ID: {info.id}")
    print(f"Service name: {info.name}")
    print(f"Service version: {info.version}")
    
    # Keep server running
    print("\nServer is running. Press Ctrl+C to stop.")
    try:
        await asyncio.Event().wait()
    except KeyboardInterrupt:
        print("\nShutting down...")
    finally:
        await service.stop()
        await nc.close()


if __name__ == "__main__":
    asyncio.run(main())
