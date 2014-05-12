#!/bin/bash
#
# You should put this bit in your ~/.bashrc
#

cred () {
    if [ -z "$GOBIN" ]; then
        echo "GOBIN is not set"
        return 1
    fi
    RES=$( $GOBIN/credulous source )
    if [ $? -eq 0 ]; then
        eval "$RES"
    else
        echo "Failed to source credentials"
        return 1
    fi
}
