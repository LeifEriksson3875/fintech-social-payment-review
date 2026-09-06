#!/bin/sh
set -eu

: "${INFRAI_API_KEY:?Set INFRAI_API_KEY before running}"
go run ./cmd/fintech-gateway
