#!/bin/bash
# post したあと、すぐに search を実行して、結果が返ってくることを確認する
set -e

echo "#11 NIP-11: Relay Information Document"
res=`curl http://localhost:9999/.well-known/nostr.json`

if echo $res | jq .software != "https://github.com/azuki774/nostar"; then
  echo "#11: ✗ Test failed"
  echo $res
  exit 1
fi

echo "#11: ✓ Test passed"
