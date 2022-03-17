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
    "add_path_components": 0,
    "blender_cmd": "{blender}",
    "blendfile": "flamenco-test.blend",
    "chunk_size": 30,
    "format": "PNG",
    "fps": 24,
    "frames": "1-60",
    "image_file_extension": ".png",
    "images_or_video": "images",
    "render_output_path": "/render/_flamenco/tests/renders/sybren/Demo for Peoples/######",
    "render_output_root": "/render/_flamenco/tests/renders/sybren/",
    "video_container_format": "MPEG1"
  },
  "priority": 50
}'
