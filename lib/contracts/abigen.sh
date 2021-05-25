#!/usr/bin/env bash

set -euo pipefail

abigen -abi "$PWD"/sfc.abi -pkg contracts -type SFC -out "$PWD/sfc_abi.go"
# abigen -abi "$PWD"/erc20.abi -pkg contracts -type ERC20 -out "$PWD/erc20_abi.go"