#!/bin/bash

DOC_TYPE=$1
TEMPLATE_PATH=$2
TARGET_PATH=$3
TITLE_PAGE=$4

cp $TEMPLATE_PATH latex-gost/tex/$1-generated.tex
if [[ -n "$TITLE_PAGE" ]]; then
	cp $TITLE_PAGE latex-gost/tex/title-page-generated.tex
fi
cd latex-gost
./build.sh $DOC_TYPE ../$TARGET_PATH
