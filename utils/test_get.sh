#!/bin/bash
T=$(python3 utils/get_test_token.py)
curl http://localhost:1234?token=$T

