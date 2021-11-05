# Publish Locks

There is a possibility to corrupt internal ZPS db by modifing its state in parallel.
For this purpose locks were introduced for publish operation.

There are following lock providers:

* dynamodb
* fs
* none

## FS locks

This method will create a file in fs with pid inside. If file doesn't exist or process with this pid already stopped, we can proceed with
modification.

```hcl
publish {
    name = "File"
    uri = "file:///tmp/zps"
    lock_uri = "file:///tmp/zps/lock"
}
```

## DynamoDB locks

Typical configuration for locks in DynamoDB will look like this. DynamoDB table should exists and have Partition Key as a key-name used in `lock_uri`.

```hcl
publish {
    name = "File"
    uri = "file:///tmp/zps-test"
    lock_uri = "dynamo://${dynamodb-table}/${key-name}"
}
```

## Manual unlock

If repository is locked and you want to unlock it manually, you can run

```bash
zps repo unlock 'Repo Name'
```
