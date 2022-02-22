#!/bin/bash

curl -v -X 'POST' \
  'http://localhost:8080/api/jobs' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "metadata": {
    "project": "Debugging Flamenco",
    "user.name": "コードモンキー"
  },
  "name": "Test Render",
  "type": "simple-blender-render",
  "settings": {
    "filepath": "flamenco-test.blend",
    "render_output": "/tmp/flamenco/test-frames",
    "chunk_size": 1,
    "extract_audio": true,
    "format": "PNG",
    "output_file_extension": ".png",
    "fps": 24,
    "frames": "1-10",
    "images_or_video": "images",
    "blender_cmd": "{blender}"
  },
  "priority": 50
}'
