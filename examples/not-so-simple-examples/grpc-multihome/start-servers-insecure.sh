#!/bin/bash

go run main.go -insecure -grpchostport ":30001" -peers localhost@localhost:30002,localhost@localhost:30003 -webport :8081&
go run main.go -insecure -grpchostport ":30002" -peers localhost@localhost:30001,localhost@localhost:30003 -webport :8082&
go run main.go -insecure -grpchostport ":30003" -peers localhost@localhost:30001,localhost@localhost:30002 -webport :8083&

