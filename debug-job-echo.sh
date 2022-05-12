#!/bin/bash

curl -X 'POST' \
  'http://localhost:8080/api/jobs' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "metadata": {
    "project": "Debugging Flamenco",
    "user.name": "dr. Sybren",
    "duration": "long"
  },
  "name": "Talk & Sleep longer",
  "priority": 3,
  "settings": {
    "sleep_duration_seconds": 20,
    "message": "Blender is {blender}"
  },
  "type": "echo-sleep-test"
}'
