#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server main.go
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"

for i in {1..3}
do
	curl "http://localhost:9999/api?key=Tom" &
done

wait
