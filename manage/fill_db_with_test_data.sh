# !/bin/bash
mkdir tests # Create tests folder
cd tests # Go inside
git clone https://github.com/kee-reel/late-sample-project # Clone sample project
cd ..
python3 fill_db.py
