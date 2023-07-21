#!/bin/bash

go build -o booking-golang cmd/web/*.go && ./booking-golang
./booking-golang -dbname=booking -dbuser=minhnguyen -cache=false -production=false -dbpass=123456