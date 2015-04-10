#!/bin/bash

if [ $1 ]; then
	tot=$1
else
	tot=1000
fi

time for (( c=0; c<=$tot; c++ )); do
	time doveadm quota recalc -u miki@tiscali.it
	if [ $? -ne 0 ]; then
		echo "ERROR: " $c
		break
	fi
done