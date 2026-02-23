#!/bin/bash
rm -rf target
mkdir -p target
cd tlcpchan && go build -o ../target/tlcpchan ./cmd/tlcpchan && cd ..
cd tlcpchan-cli && go build -o ../TLCPCHAN ../target/tlcpchan-cli && cd ..
cd tlcpchan-ui && npm run build && cp -r dist ../target/ui && cd ..
