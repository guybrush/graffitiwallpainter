#!/bin/bash
IMG="guybrush/graffitiwallpainter"
docker build -t $IMG .
docker push $IMG
