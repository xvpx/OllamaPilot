#!/bin/bash

echo "🔄 Rebuilding Docker containers..."
docker-compose down
docker-compose build --no-cache api
docker-compose up -d

echo "⏳ Waiting for services to start..."
sleep 10

echo "🔍 Testing endpoints..."
echo "Testing health endpoint:"
curl -s http://localhost:8080/health | jq .

echo -e "\nTesting v1 test endpoint:"
curl -s http://localhost:8080/v1/test | jq .

echo -e "\nTesting v1 chat endpoint with POST:"
curl -s -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "test", "session_id": "test-session", "model": "llama2:7b", "stream": false}' | jq .

echo -e "\n📋 Container logs (last 20 lines):"
docker-compose logs --tail=20 api