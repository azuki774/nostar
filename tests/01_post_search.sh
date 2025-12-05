#!/bin/bash
# post したあと、すぐに search を実行して、結果が返ってくることを確認する
set -e

echo "#01-1: Post powa"
algia powa

echo "#01-2: search powa"
algia search | grep 'ぽわ〜'
