#!/bin/bash

curl -v -X 'POST' \
  'http://localhost:8080/api/v3/jobs' \
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
    "add_path_components": 0,
    "blender_cmd": "{blender}",
    "blendfile": "flamenco-test.blend",
    "chunk_size": 30,
    "format": "PNG",
    "fps": 24,
    "frames": "1-60",
    "image_file_extension": ".png",
    "images_or_video": "images",
    "render_output_path": "/tmp/flamenco/Demo for Peoples/######",
    "render_output_root": "/tmp/flamenco/",
    "video_container_format": "MPEG1"
  },
  "priority": 50,
  "submitter_platform": "manager"
}'
