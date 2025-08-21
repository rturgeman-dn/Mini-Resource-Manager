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

echo "C5. Release not allocated (try to release 100 again if free)"
curl -sX POST http://localhost:8080/release -H "Content-Type: application/json" -d '{"pool":"vlan-pool","value":100}'
echo -e "\n"

echo "=== DONE ==="
