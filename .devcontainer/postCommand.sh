#!/bin/bash

export PGPASSWORD=postgres

## マイグレーション
psql -h localhost -U postgres -c "CREATE DATABASE nostar;"
psql -h localhost -U postgres -d nostar -f scripts/migrations/001_init.sql
