#!/bin/bash
EMAIL=admin@admin.com
python3 utils/add_user.py admin $EMAIL
python3 utils/create_token.py $EMAIL 1
