## Building

1. Setup environment variables

    ```bash
    mkdir lib/{bin,pkg,src}
    export GOPATH=$(pwd)/lib

2. Install deps
   
    ```bash
    go install github.com/aws/aws-sdk-go@latest
    ```

3. Build

    ```bash
    go build -o app .
    chmod +x ./app
    ```

4. Run

    ```bash
    ./app -role-arn <aws-role-arn> -region <aws-region> -param-path <aws-ssm-param-path>
    ```

    The application will expect to find AWS credentials set from ENV, so ensure that the IAM user credentials have been set up correctly, and if necessary, export the correct `AWS_PROFILE` variable as well.

## Testing

To test the application, set up the IAM role as before, then run the application directly without building:

```bash
go run app.go -role-arn <aws-role-arn> -region <aws-region> -param-path <aws-ssm-param-path> -key-path <some-local-path-for-testing>
```
