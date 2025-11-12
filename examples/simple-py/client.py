"""
Simple NATS Micro client example using protoc-gen-nats-micro
"""

import asyncio
import sys
from pathlib import Path

# Add generated code to path
sys.path.insert(0, str(Path(__file__).parent / "gen"))

import nats
from example.v1.service_nats_pb2 import ExampleServiceClient
from example.v1 import service_pb2 as pb


async def main():
    """Main client function"""
    # Connect to NATS
    nc = await nats.connect("nats://localhost:4222")
    print("Connected to NATS")
    
    # Create client
    client = ExampleServiceClient(nc)
    print("ExampleService client created")
    
    # Test Echo
    print("\n=== Testing Echo ===")
    echo_req = pb.EchoRequest(message="Hello from Python!")
    try:
        echo_resp, headers = await client.echo(echo_req)
        print(f"Response: {echo_resp.message}")
        print(f"Timestamp: {echo_resp.timestamp}")
        print(f"Response headers: {headers}")
    except Exception as e:
        print(f"Error: {e}")
    
    # Test Echo with custom headers
    print("\n=== Testing Echo with Headers ===")
    try:
        echo_resp, headers = await client.echo(
            echo_req,
            headers={"X-User-ID": "12345", "X-Request-ID": "abc-def"}
        )
        print(f"Response: {echo_resp.message}")
        print(f"Response headers: {headers}")
    except Exception as e:
        print(f"Error: {e}")
    
    # Test GetGreeting
    print("\n=== Testing GetGreeting ===")
    greeting_req = pb.GetGreetingRequest(name="Python Developer")
    try:
        greeting_resp, headers = await client.get_greeting(greeting_req)
        print(f"Greeting: {greeting_resp.greeting}")
        print(f"Response headers: {headers}")
    except Exception as e:
        print(f"Error: {e}")
    
    # Test timeout
    print("\n=== Testing Timeout ===")
    try:
        greeting_resp, headers = await client.get_greeting(
            greeting_req,
            timeout=0.001  # Very short timeout
        )
        print(f"Greeting: {greeting_resp.greeting}")
    except Exception as e:
        print(f"Timeout error (expected): {e}")
    
    # Show endpoints
    print("\n=== Available Endpoints ===")
    for endpoint in client.endpoints():
        print(f"  {endpoint}")
    
    await nc.close()
    print("\nClient done!")


if __name__ == "__main__":
    asyncio.run(main())
