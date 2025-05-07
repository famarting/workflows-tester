# Workflow tester

## Java
```
mvn clean install
diagrid dev run -a app1 -- java -jar target/workflow-app-0.0.1-SNAPSHOT.jar --port=5001
```

then
```
diagrid dev run -a app1 -- go run ./tester/main.go
```

## Golang
```
diagrid dev run -a app1 -- go run ./workflow/main.go
```

then
```
WF_NAME=OrderProcessingWorkflow diagrid dev run -a app1 -- go run ./tester/main.go
```
