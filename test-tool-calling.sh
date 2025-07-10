#!/bin/bash
curl -X POST http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1-gpu",
    "messages": [
      {
        "role": "user",
        "content": "Write a file named test.txt with the content Hello World"
      }
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "write",
          "description": "Write content to a file",
          "parameters": {
            "type": "object",
            "properties": {
              "filename": {
                "type": "string",
                "description": "Name of the file to write"
              },
              "content": {
                "type": "string",
                "description": "Content to write to the file"
              }
            },
            "required": ["filename", "content"]
          }
        }
      }
    ],
    "tool_choice": "auto"
  }'
