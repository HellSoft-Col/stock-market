#!/usr/bin/env python3
"""
Simple connection test script
=============================

Test basic connectivity to the trading server before running full simulation.
"""

import asyncio
import json
import sys
import time

try:
    import websockets
except ImportError:
    print("Error: websockets library not installed")
    print("Install with: pip install websockets")
    sys.exit(1)

async def test_connection(server_url="ws://localhost:8080", token="TK-TEST"):
    """Test basic connection and authentication"""
    print(f"Testing connection to {server_url} with token {token}")
    
    try:
        # Connect to server
        print("Connecting...")
        websocket = await websockets.connect(server_url)
        print("✓ Connected successfully")
        
        # Test authentication
        auth_message = {
            "type": "AUTH",
            "data": {"token": token}
        }
        await websocket.send(json.dumps(auth_message))
        print("✓ Authentication message sent")
        
        # Wait for response
        response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
        auth_response = json.loads(response)
        print(f"✓ Received response: {auth_response}")
        
        if auth_response.get("type") == "AUTH_SUCCESS":
            print("✓ Authentication successful!")
            
            # Test a simple ping
            ping_message = {"type": "PING", "data": {}}
            start_time = time.time()
            await websocket.send(json.dumps(ping_message))
            
            pong_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
            response_time = time.time() - start_time
            print(f"✓ Ping response time: {response_time:.3f}s")
            
        else:
            print(f"✗ Authentication failed: {auth_response}")
            return False
        
        await websocket.close()
        print("✓ Connection closed gracefully")
        return True
        
    except asyncio.TimeoutError:
        print("✗ Timeout waiting for server response")
        return False
    except ConnectionRefusedError:
        print("✗ Connection refused - is the server running?")
        return False
    except Exception as e:
        print(f"✗ Connection failed: {e}")
        return False

async def main():
    if len(sys.argv) > 1:
        server_url = sys.argv[1]
    else:
        server_url = "ws://localhost:8080"
    
    if len(sys.argv) > 2:
        token = sys.argv[2]
    else:
        token = "TK-TEST"
    
    success = await test_connection(server_url, token)
    
    if success:
        print("\n🎉 Server connection test PASSED!")
        print("You can now run the full simulation.")
        return 0
    else:
        print("\n❌ Server connection test FAILED!")
        print("Please check that the server is running and accessible.")
        return 1

if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)