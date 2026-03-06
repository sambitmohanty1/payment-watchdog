#!/bin/bash
for d in api worker; do
  cd $d
  GOTOOLCHAIN=local go fmt ./...
  cd ..
done
