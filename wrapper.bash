#!/bin/bash
#
# You should put this bit in your ~/.bashrc
#

cred () {
    RES=$( credulous $@)
    if [ $? -eq 0 ]; then
        echo -n "Loading AWS creds into current environment..."
        eval "$RES"
        if [ $? -eq 0 ]; then
          echo "OK"
        else
          echo "FAIL"
        fi
    else
        echo "Failed to source credentials"
        return 1
    fi
}
