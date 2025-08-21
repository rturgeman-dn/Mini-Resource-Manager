#!/bin/bash

echo "=== A) Setup ==="

echo "A1. Create 'vlan' template"
curl -sX POST http://localhost:8080/templates -H "Content-Type: application/json" -d '{"name":"vlan", "min":100, "max":105}'
echo -e "\n"

echo "A2. Create pool 'vlan-pool'"
curl -sX POST http://localhost:8080/pools -H "Content-Type: application/json" -d '{"name":"vlan-pool", "template":"vlan"}'
echo -e "\n"

echo "=== B) Happy path allocations & release ==="

echo "B1. Allocate 1"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"

echo "B2. Allocate 2"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"

echo "B3. Allocate 3"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"

echo "B4. Release 102"
curl -sX POST http://localhost:8080/release -H "Content-Type: application/json" -d '{"pool":"vlan-pool", "value":102}'
echo -e "\n"

echo "B5. Allocate 2 more (expect 102 and 103)"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"

echo "B6. Allocate until exhaustion (104, 105, then error)"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"vlan-pool"}'
echo -e "\n"

echo "=== C) Errors ==="

echo "C1. Invalid template (min > max)"
curl -sX POST http://localhost:8080/templates -H "Content-Type: application/json" -d '{"name":"bad","min":7,"max":3}'
echo -e "\n"

echo "C2. Duplicate template"
curl -sX POST http://localhost:8080/templates -H "Content-Type: application/json" -d '{"name":"vlan","min":1,"max":2}'
echo -e "\n"

echo "C3. Missing pool in allocate"
curl -sX POST http://localhost:8080/allocate -H "Content-Type: application/json" -d '{"pool":"nope"}'
echo -e "\n"

echo "C4. Release out of range"
curl -sX POST http://localhost:8080/release -H "Content-Type: application/json" -d '{"pool":"vlan-pool","value":999}'
echo -e "\n"

echo "C5. Release not allocated (try to release 100 need to get released and then not_allocated)"
curl -sX POST http://localhost:8080/release -H "Content-Type: application/json" -d '{"pool":"vlan-pool","value":100}'
echo -e "\n"
curl -sX POST http://localhost:8080/release -H "Content-Type: application/json" -d '{"pool":"vlan-pool","value":100}'
echo -e "\n"


echo "=== D1. Parallel allocate (fan-in) ==="

# Step 1: Ensure the pool is clean
echo "- Creating template and pool"
curl -sX POST http://localhost:8080/templates \
  -H "Content-Type: application/json" \
  -d '{"name":"new-vlan","min":100,"max":105}'

curl -sX POST http://localhost:8080/pools \
  -H "Content-Type: application/json" \
  -d '{"name":"vlan-pool-d1","template":"new-vlan"}'

# Step 2: Fire 8 parallel allocations
echo "- Sending 8 parallel allocate requests..."
for i in {1..8}; do
  curl -sX POST http://localhost:8080/allocate \
    -H "Content-Type: application/json" \
    -d '{"pool":"vlan-pool-d1"}' &
done

wait

echo "- Done. Expected: 6 values (100–105) + 2 errors (no_free_items)"

echo "=== D2. Backpressure/timeout ==="

# Step 1: Create template and pool
echo "- Creating template and pool with small buffer size simulation"
curl -sX POST http://localhost:8080/templates -H "Content-Type: application/json" -d '{"name":"tiny", "min":100, "max":105}'
curl -sX POST http://localhost:8080/pools -H "Content-Type: application/json" -d '{"name":"tiny-pool", "template":"tiny"}'

echo "- Sending 50 parallel allocate requests with 500ms timeout..."
for i in {1..50}; do
  curl -sX POST http://localhost:8080/allocate     -H "Content-Type: application/json"     -m 0.5     -d '{"pool":"tiny-pool"}' &
done

wait
echo "- Done. Expect mix of 200 OK and 408 Request Timeout."


echo "=== DONE ==="
