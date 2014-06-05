#!/bin/bash
#
# credulous.sh
#
# Wrapper that sources output of 'credulous source' command into
# current environment
#
# Source this from your ~/.bash_profile

credulous () {
    BINARY=$( type -P credulous )
    RES=$( $BINARY $@ )
    RET=$?
    if [ $RET -eq 0 -a "x$1" = "xsource" ]; then
        echo -n "Loading AWS creds into current environment..."
        eval "$RES"
        if [ $? -eq 0 ]; then
          echo "OK"
        else
          echo "FAIL"
        fi
    elif [ $RET -eq 0 ]; then
        echo "$RES"
    else
        echo "Failed to source credentials"
        return 1
    fi
}
