#!/bin/bash

curl -X 'POST' \
  'http://localhost:8080/api/jobs' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "metadata": {
    "project": "Debugging Flamenco",
    "user.name": "コードモンキー"
  },
  "name": "Talk & Sleep",
  "priority": 50,
  "settings": {
    "sleep_duration_seconds": 3,
    "message": "{blender}"
  },
  "type": "echo-sleep-test"
}'
