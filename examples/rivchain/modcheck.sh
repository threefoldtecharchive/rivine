#!/bin/bash

if [[ $(git status --porcelain) ]] ; then
    echo "Files have been modified by RivineCG" && exit 1
else
    echo "No new modifications made by RivineCG, all good"
fi;
