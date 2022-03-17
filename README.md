# datasync-api

This API provides simple restful API access to the AWS DataSync service.

## Endpoints

```
GET /v1/test/ping
GET /v1/test/version
GET /v1/test/metrics

GET    /v1/datasync/{account}/movers
POST   /v1/datasync/{account}/movers/{group}
GET    /v1/datasync/{account}/movers/{group}
GET    /v1/datasync/{account}/movers/{group}/{id}
POST   /v1/datasync/{account}/movers/{group}/{name}/{start}
GET	   /v1/datasync/{account}/movers/{group}/{name}/{runs}
GET	   /v1/datasync/{account}/movers/{group}/{name}/runs/{id}
```

## Authentication

Authentication is accomplished via an encrypted pre-shared key in the `X-Auth-Token` header.

## Usage

### Create Data Mover

Create requests are asynchronous and return a task ID in the header `X-Flywheel-Task`. This header can be used to get the task information and logs from the flywheel HTTP endpoint.

POST `/v1/datasync/{account}/movers/{group}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **202 Acepted**               | creating a data mover           |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example create request body

```json
{
    "Name": "best-effort-datasync-01",
    "Source": {
        "Type": "S3",
        "S3": {
            "S3BucketArn": "arn:aws:s3:::tester1234567890.example.com",
            "S3StorageClass": "STANDARD",
            "Subdirectory": "/"
        }
    },
	"Destination": {
        "Type": "S3",
        "S3": {
            "S3BucketArn": "arn:aws:s3:::receiver1234567890.example.com",
            "S3StorageClass": "STANDARD",
            "Subdirectory": "/"
        }
    },
	"Tags": [
        {
            "Key": "env",
            "Value": "sbx"
        }
    ]
}
```

#### Example create response headers

```json
{
    "X-Flywheel-Task": "3b9ee9e9-9ffa-4b07-93c7-59e7e6f1fa7f"
}
```

### List all Data Movers

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
    "best-effort-datasync-01",
    "latasync-2021",
    "from-here-to-there-2"
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
    "best-effort-datasync-01",
    "latasync-2021"
]
```

### Get details about a Data Mover, including its task, source and destination locations

GET `/v1/datasync/{account}/movers/{group}/{name}`

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
        "Name": "best-effort-datasync-01",
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

### Delete Data Mover

DELETE `/v1/datasync/{account}/movers/{group}/{name}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **204 No Content**            | deleted the data mover          |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

### Run a Data Mover Task

POST `/v1/datasync/{account}/movers/{group}/{name}/{start}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | start the data mover            |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example list response

```json
[
    "TaskExecutionArn": "arn:aws:datasync:us-east-1:516855177326:task/task-05cd6f77d7b5d15ac/execution/exec-086d6c629a6bf3581"
]
```

### List All Data Mover Runs

GET `/v1/datasync/{account}/movers/{group}/{name}/{runs}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | list data mover runs            |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example list response

```json
[
    "exec-0816cd98c4791fb39",

    "exec-00d17529fe536568f",

    "exec-0cd145fd8fd6c751d",

    "exec-0668badbcf3ba3f1b",

    "exec-086d6c629a6bf3581"
]
```

### Get Information about a Data Mover Run

GET `/v1/datasync/{account}/movers/{group}/{name}/{runs}/{id}`

| Response Code                 | Definition                      |
| ----------------------------- | --------------------------------|
| **200 OK**                    | get data mover run information  |
| **400 Bad Request**           | badly formed request            |
| **404 Not Found**             | account not found               |
| **500 Internal Server Error** | a server error occurred         |

#### Example list response

```json
{
    "BytesTransferred": 0,
    "BytesWritten": 0,
    "EstimatedBytesToTransfer": 0,
    "EstimatedFilesToTransfer": 0,
    "FilesTransferred": 0,
    "StartTime": "2022-03-01T14:04:17.986Z",
    "Status": "SUCCESS",
    "Result": {
        "ErrorCode": null,
        "ErrorDetail": null,
        "PrepareDuration": 1669,
        "PrepareStatus": "SUCCESS",
        "TotalDuration": 7662,
        "TransferDuration": 5870,
        "TransferStatus": "SUCCESS",
        "VerifyDuration": 155,
        "VerifyStatus": "SUCCESS"
    }
}
```
## License

GNU Affero General Public License v3.0 (GNU AGPLv3)  
Copyright © 2021 Yale University
