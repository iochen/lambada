# lambada
A pipe for HTTP(s) traffics based on serverless functions.

```
+-------------------------------------------------------+  
|  This application can tranfer HTTP(s) traffics ONLY!  |  
+-------------------------------------------------------+  
```

### Lambda
1. Create a **aws lambda function** with ***Runtime*** set to **`Go 1.x`**.
2. In the ***Code*** tab of your created lambda function, upload source from zipped binary file built from [cmd/server/lambda](https://github.com/iochen/lambada/tree/master/cmd/server/lambda).
3. Set ***Handler*** to the name of your built binary file name.
4. Add `LBD_KEY` and `LBD_USER` environment variables in lambda functions/
5. Create an api-gateway connected to your function when receiving **POST** request.
6. Find server url from api-gateway.

### Client
1. Build binary from [cmd/client](https://github.com/iochen/lambada/tree/master/cmd/client)
2. Generate a CA key pair. (You can use *`openssl`* or *`v2ctl`*)
3. Edit `config.yaml` to your own settings.
