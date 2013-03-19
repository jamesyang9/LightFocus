#!/bin/bash

for file in *.m4v 
do 
    fname=`basename $file .m4v`
    echo $fname
    mkdir $fname\_temp
    chmod 755 $fname\_temp
    echo ffmpeg -i $fname -r 30 -f image2 $fname\_temp/%2d.png
    ffmpeg -i $file -r 30 -f image2 $fname\_temp/%2d.png

done