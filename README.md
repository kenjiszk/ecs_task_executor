# ecs_task_executor
## What is this?
This is wrapper of aws sdk for ecs task execution. `ecs-cli` can execute ECS Task by using docker-compose.yml, but it executes a task in asynchronous. `ecs_task_executor` can execute task in synchronous.

## Samples.
### Before executing ecs_task_executor
#### Prepare docker-compose.yml
```
version: 2
services:
  web:
    image: 'ubuntu:16.04'
    command: sample_task.sh
    logging:
      driver: "awslogs"
      options:
        awslogs-region: "ap-northeast-1"
        awslogs-group: "kenjiszk-test"
        awslogs-stream-prefix: "kenjiszk-test"
```
#### Create new task definition by ecs-cli
```
$ ecs-cli compose --project-name kenjiszk-test --file docker-compose.yml create
INFO[0000] Using ECS task definition                     TaskDefinition="kenjiszk-test:1"
```
#### Set environment variables
```
$ export AWS_DEFAULT_REGION=ap-northeast-1
$ export AWS_ACCESS_KEY_ID=XXXXXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXX
```
### Success case.
```
$ ecs_task_executor --cluster Sample -t kenjiszk-test:1 -n web -c 'sleep 30'
Set timeout as 600 sec.
LastStatus=PENDING TimeElapsed=5.017381771s
LastStatus=PENDING TimeElapsed=10.037772876s
LastStatus=PENDING TimeElapsed=15.051642439s
LastStatus=RUNNING TimeElapsed=20.072351989s
LastStatus=RUNNING TimeElapsed=25.087819946s
LastStatus=RUNNING TimeElapsed=30.103733236s
LastStatus=RUNNING TimeElapsed=35.125751885s
LastStatus=RUNNING TimeElapsed=40.144906234s
LastStatus=RUNNING TimeElapsed=45.169059897s
LastStatus=STOPPED TimeElapsed=45.184647817s
Task successfully finished.
```
### Fail case. command is bad.
```
$ ecs_task_executor --cluster Sample -t kenjiszk-test:1 -n web -c 'not_found_command'
Set timeout as 600 sec.
LastStatus=PENDING TimeElapsed=5.022523186s
LastStatus=PENDING TimeElapsed=10.03916485s
{
  ContainerArn: "arn:aws:ecs:ap-northeast-1:00000000000:container/5cc28647-13f9-469d-881d-8217c393eebb",
  HealthStatus: "UNKNOWN",
  LastStatus: "STOPPED",
  Name: "web",
  NetworkInterfaces: [],
  Reason: "CannotStartContainerError: API error (400): OCI runtime create failed: container_linux.go:296: starting container process caused \"exec: \\\"not_found_command\\\": executable file not found in $PATH\": unknown\n",
  TaskArn: "arn:aws:ecs:ap-northeast-1:00000000000:task/f8c0fe8b-2d1e-49f6-ab42-1d4d3d337c86"
}
exit status 1
```
### Fail case. command is good, but failed.
Sample script `run_and_fail.sh`. It will be successfully executed, but failed after 30min. 
```
#!/bin/bash

sleep 30;
exit 255;
```
```
$ ecs_task_executor --cluster Sample -t kenjiszk-test:1 -n web -c 'run_and_fail.sh'
Set timeout as 600 sec.
LastStatus=PENDING TimeElapsed=5.024269701s
LastStatus=PENDING TimeElapsed=10.045879005s
LastStatus=PENDING TimeElapsed=15.070078421s
LastStatus=RUNNING TimeElapsed=20.092714283s
LastStatus=RUNNING TimeElapsed=25.109575722s
LastStatus=RUNNING TimeElapsed=30.131358467s
LastStatus=RUNNING TimeElapsed=35.150669821s
LastStatus=RUNNING TimeElapsed=40.16816051s
LastStatus=RUNNING TimeElapsed=45.185257392s
LastStatus=RUNNING TimeElapsed=50.209708875s
{
  ContainerArn: "arn:aws:ecs:ap-northeast-1:000000000:container/df87b9cd-962f-4580-ad8d-f6b97446b9a2",
  ExitCode: 255,
  HealthStatus: "UNKNOWN",
  LastStatus: "STOPPED",
  Name: "web",
  NetworkBindings: [],
  NetworkInterfaces: [],
  TaskArn: "arn:aws:ecs:ap-northeast-1:00000000:task/de79a37c-4c51-4307-826c-f6a3c7d733cb"
}
exit status 255
```
