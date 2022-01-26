#!/bin/bash
T=$(python3 get_test_token.py)
curl http://web:1234?token=$T

