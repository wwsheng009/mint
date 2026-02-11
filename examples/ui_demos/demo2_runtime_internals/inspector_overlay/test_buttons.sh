#!/bin/bash
echo "Running inspector demo for 5 seconds..."
timeout 5 ./main.exe &
sleep 5
echo "Checking logs..."
if [ -f logs/application.log ]; then
    echo "=== Found enrichHitMap logs ==="
    grep "enrichHitMap" logs/application.log | tail -20
    echo "=== Found handleMsg logs ==="  
    grep "handleMsg" logs/application.log | tail -20
    echo "=== Found Instance.Handle logs ==="
    grep "Instance.Handle" logs/application.log | tail -20
else
    echo "No log file found - app may have crashed or not run"
fi
