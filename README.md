# datasync-api

This API provides simple restful API access to the AWS DataSync service.

## Endpoints

```
GET /v1/test/ping
GET /v1/test/version
GET /v1/test/metrics

GET    /v1/datasync/{account}/movers
GET    /v1/datasync/{account}/movers/{group}
GET    /v1/datasync/{account}/movers/{group}/{id}
```

## Authentication

Authentication is accomplished via an encrypted pre-shared key in the `X-Auth-Token` header.

## Usage

### List Data Movers

GET `/v1/datasync/{account}/movers`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | return the list of data movers  |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example list response

```json
[
    "task-0ee229c003ef4f0b8",
    "task-06c39db367408c721"
    "task-00261ad4da6768578"
]
```

### List Data Movers by group id

GET `/v1/datasync/{account}/movers/{group}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | return the list of data movers  |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example list by group response

```json
[
    "task-0ee229c003ef4f0b8",
    "task-00261ad4da6768578"
]
```

### Get details about a Data Mover, including its task, source and destination locations

GET `/v1/datasync/{account}/movers/{group}/{id}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | return details of a data mover  |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account or mover not found      |
| **500 Internal Server Error** | a server error occurred         |

#### Example show response

```json
{
    "Task": {
        "CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:1234567890:log-group:/aws/datasync",
        "CreationTime": "2021-12-01T21:18:40.546Z",
        "CurrentTaskExecutionArn": null,
        "DestinationLocationArn": "arn:aws:datasync:us-east-1:1234567890:location/loc-0379907909ce0b7d9",
        "DestinationNetworkInterfaceArns": [
            "arn:aws:ec2:us-east-1:1234567890:network-interface/eni-05c0bd4389d26aaae",
            "arn:aws:ec2:us-east-1:1234567890:network-interface/eni-010933e71d6a33451",
            "arn:aws:ec2:us-east-1:1234567890:network-interface/eni-05145e6e6ecce5ce7",
            "arn:aws:ec2:us-east-1:1234567890:network-interface/eni-066fee4100f09ef44"
        ],
        "ErrorCode": null,
        "ErrorDetail": null,
        "Excludes": [],
        "Includes": [],
        "Name": "tgtest3",
        "Options": {
            "Atime": "BEST_EFFORT",
            "BytesPerSecond": -1,
            "Gid": "NONE",
            "LogLevel": "OFF",
            "Mtime": "PRESERVE",
            "OverwriteMode": "ALWAYS",
            "PosixPermissions": "NONE",
            "PreserveDeletedFiles": "PRESERVE",
            "PreserveDevices": "NONE",
            "SecurityDescriptorCopyFlags": "NONE",
            "TaskQueueing": "ENABLED",
            "TransferMode": "CHANGED",
            "Uid": "NONE",
            "VerifyMode": "ONLY_FILES_TRANSFERRED"
        },
        "Schedule": null,
        "SourceLocationArn": "arn:aws:datasync:us-east-1:1234567890:location/loc-0126cee0d76502bb1",
        "SourceNetworkInterfaceArns": [],
        "Status": "AVAILABLE",
        "TaskArn": "arn:aws:datasync:us-east-1:1234567890:task/task-0ee229c003ef4f0b8"
    },
    "Source": {
        "SMB": {
            "AgentArns": [
                "arn:aws:datasync:us-east-1:1234567890:agent/agent-0914d8e6e0674c8b7"
            ],
            "CreationTime": "2021-11-01T20:35:44.471Z",
            "LocationArn": "arn:aws:datasync:us-east-1:1234567890:location/loc-0126cee0d76502bb1",
            "LocationUri": "smb://storage.example.com/home/",
            "MountOptions": {
                "Version": "AUTOMATIC"
            },
            "User": "tester"
        }
    },
    "Destination": {
        "S3": {
            "AgentArns": null,
            "CreationTime": "2021-11-01T20:35:44.362Z",
            "LocationArn": "arn:aws:datasync:us-east-1:1234567890:location/loc-0379907909ce0b7d9",
            "LocationUri": "s3://tester1234567890.example.com/",
            "S3Config": {
                "BucketAccessRoleArn": "arn:aws:iam::1234567890:role/service-role/AWSDataSyncS3BucketAccess-tester1234567890.example.com"
            },
            "S3StorageClass": "STANDARD"
        }
    },
    "Tags": [
        {
            "Key": "spinup:flavor",
            "Value": "datamover"
        },
        {
            "Key": "spinup:org",
            "Value": "spindev"
        },
        {
            "Key": "spinup:spaceid",
            "Value": "abc-123"
        },
        {
            "Key": "spinup:type",
            "Value": "storage"
        }
    ]
}
```

## License

GNU Affero General Public License v3.0 (GNU AGPLv3)  
Copyright © 2021 Yale University
