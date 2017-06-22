#!/bin/bash

go run main.go -p $1 -webport :8081&
go run main.go -p $1 -webport :8082&
go run main.go -p $1 -webport :8083&

