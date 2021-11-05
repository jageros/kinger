#!/bin/bash
command="protoc3 --gofast_out=proto/pb proto/pbdef/*.proto --proto_path=proto/pbdef"
echo $command
`$command`

python tools/gen_meta.py
