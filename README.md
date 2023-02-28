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