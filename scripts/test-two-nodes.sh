#!/bin/bash

echo "Starting Node A (Bootstrap)..."
./bin/dvpnd -config configs/node-a.yaml &
NODE_A_PID=$!

sleep 3

echo "Starting Node B..."
./bin/dvpnd -config configs/node-b.yaml &
NODE_B_PID=$!

sleep 3

echo "=== Node Status ==="
echo "Node A status:"
curl -s http://127.0.0.1:8334/api/status | jq

echo -e "\nNode B status:"
curl -s http://127.0.0.1:8435/api/status | jq

echo -e "\n=== Connecting Nodes ==="
echo "Connecting Node B to Node A..."
curl -s -X POST http://127.0.0.1:8435/api/peers/connect \
     -H "Content-Type: application/json" \
     -d '{"address":"127.0.0.1:8333"}' | jq

sleep 2

echo -e "\n=== Post-Connection Status ==="
echo "Node A peers:"
curl -s http://127.0.0.1:8334/api/peers | jq

echo "Node B peers:"
curl -s http://127.0.0.1:8435/api/peers | jq

echo -e "\n=== Testing Mining on Node A ==="
curl -s -X POST http://127.0.0.1:8334/api/mining/start | jq

sleep 5

echo "Node A mining status:"
curl -s http://127.0.0.1:8334/api/mining/status | jq

echo -e "\nPress any key to stop nodes..."
read -n 1

echo "Stopping nodes..."
kill $NODE_A_PID $NODE_B_PID 2>/dev/null
wait $NODE_A_PID $NODE_B_PID 2>/dev/null

echo "Test complete!"
