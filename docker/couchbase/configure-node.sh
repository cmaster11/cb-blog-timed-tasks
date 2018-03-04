#!/bin/bash

set -x
set -m

/entrypoint.sh couchbase-server &

# Waits for CB to startup
sleep 15

# Setup index and memory quota
curl -v -X POST http://127.0.0.1:8091/pools/default -d memoryQuota=1024 -d indexMemoryQuota=512

# Setup services
curl -v http://127.0.0.1:8091/node/controller/setupServices -d services=kv%2Cn1ql%2Cindex

# Setup credentials
curl -v http://127.0.0.1:8091/settings/web -d port=8091 -d username=Administrator -d password=password

# Setup Memory Optimized Indexes
curl -i -u Administrator:password -X POST http://127.0.0.1:8091/settings/indexes -d 'storageMode=memory_optimized'

curl -i -X POST -u Administrator:password \
  -d 'name=timed_tasks' \
  -d 'authType=sasl' \
  -d 'saslPassword=bucket' \
  -d 'bucketType=couchbase' \
  -d 'ramQuotaMB=768' \
  -d 'replicaNumber=0' \
  -d 'proxyPort=30001' \
  http://127.0.0.1:8091/pools/default/buckets

# Create user
couchbase-cli user-manage -c 127.0.0.1:8091 -u Administrator \
 -p password --set --rbac-username 'timed_tasks' --rbac-password 'timed_tasks_pwd' \
 --rbac-name "Timed tasks user" --roles 'bucket_full_access[timed_tasks],query_delete[timed_tasks],query_select[timed_tasks],query_insert[timed_tasks],query_update[timed_tasks]' \
 --auth-domain local

fg 1

