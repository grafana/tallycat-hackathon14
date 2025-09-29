#!/bin/bash


cleanup() {
    echo "Shutting down services..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    wait $BACKEND_PID $FRONTEND_PID 2>/dev/null
    exit 0
}


trap cleanup SIGTERM SIGINT


echo "Starting backend server..."
cd /app
go run main.go server &
BACKEND_PID=$!


sleep 3


echo "Starting frontend server..."
cd /app/ui
npm run dev &
FRONTEND_PID=$!


echo "Both servers started. Backend PID: $BACKEND_PID, Frontend PID: $FRONTEND_PID"
echo "Frontend available at: http://localhost:4000"


wait