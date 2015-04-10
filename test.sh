#!/bin/bash

opt()
{
    case "$1" in
        -c)
            shift
            TOT=$1
            shift
            opt $*
            ;;
        -cmd)
            shift
            CMD=$1
            shift
            opt $*
            ;;
		-u)
			shift
			USER=$1
			shift
			opt $*
			;;
		-h)
			echo "-c <iterations>"
			echo "-u <username>"
			echo "-cmd <command> [get|recalc]"
			exit 0
			;;
        *)
            RETURN=${*}
            ;;	            
    esac
}

USER=miki
CMD=get
TOT=100

opt $*

time for (( c=0; c<=${TOT}; c++ )); do
	doveadm quota ${CMD} -u ${USER}@tiscali.it
	if [ $? -ne 0 ]; then
		echo "ERROR: " $c
		break
	fi
done