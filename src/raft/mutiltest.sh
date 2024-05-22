#!/bin/bash

fail_count=0
thread_count=100
test_count=1000

for ((i=0; i<$test_count; i++)); do
    ((index=i%thread_count))
    {
        go test -run 3A > "out_$i"
        if [ $? -ne 0 ]; then
            mv "out_$i" "fail_out_$i"
            ((fail_count++))
        else
            rm "out_$i"
        fi
    } &
    if (( (i+1)%thread_count == 0 )); then
        wait
    fi
done

wait

echo "Number of failed tests: $fail_count"