#!/bin/bash
cd ~/oj/backend
export $(cat .env 2>/dev/null | xargs)
go run ./cmd/server
