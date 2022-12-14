#!/bin/bash

curl -X 'POST' \
  'http://localhost:8080/api/v3/jobs' \
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
    "sleep_repeats": 1,
    "message": "Blender is {blender}"
  },
  "type": "echo-sleep-test",
  "submitter_platform": "manager"
}'
