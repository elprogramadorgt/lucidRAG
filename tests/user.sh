#!/bin/bash


curl -X POST http://localhost:8080/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Edu",
    "last_name": "elprogramadorgt",
    "email": "barney@elprogramadorgt.fun",
    "password": "admin"
    }'