* create root ca

    `pki_create_ca`:

    ```
    $ ziti pki create ca --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-file=root-ca --ca-name="zrok.io Root CA"
    ```

* signing root ca

    `pki_create_ca`:

    ```
    $ ziti pki create ca --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-file=signing-root-ca --ca-name="zrok.io Signing Root CA"
    ```

* intermediate

    `pki_create_intermediate`:

    ```
    $ ziti pki create intermediate --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=root-ca --intermediate-name="zrok.io Intermediate" --intermediate-file=intermediate --max-path-len=1
    ```

* signing intermediate

    `pki_create_intermediate`:

    ```
    $ ziti pki create intermediate --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=intermediate --intermediate-name="zrok.io Signing Intermediate" --intermediate-file=signing-intermediate --max-path-len=1
    ```

* create controller client/server certs:

    `pki_client_server`:

    ```
    $ ziti pki create server --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=intermediate --server-file=ctrl-server --dns="ziti.dev.zrok.io,localhost" --ip="0.0.0.0,10.0.0.41,127.0.01" --server-name="zrok.io controller server"
    $ ziti pki create client --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=intermediate --client-file=ctrl-client --key-file=ctrl-server --client-name="zrok.io controller client"
    ```

* create edge router client/server certs:

    `pki_client_server`:

    ```
    $ ziti pki create server --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=intermediate --server-file=router0-server --dns="ziti.dev.zrok.io,localhost" --ip="0.0.0.0,10.0.0.41,127.0.01" --server-name="zrok.io router0 server"
    $ ziti pki create client --pki-root=/home/ubuntu/local/etc/zrok.io/pki --ca-name=intermediate --client-file=router0-client --key-file=router0-server --client-name="zrok.io router0 client"
    ```

* `cas.pem`:

    `createControllerConfig`:

    ```
    $ cat local/etc/zrok.io/pki/intermediate/certs/ctrl-server.chain.pem > local/etc/zrok.io/pki/cas.pem
    $ cat local/etc/zrok.io/pki/intermediate/certs/signing-intermediate.cert >> local/etc/zrok.io/pki/cas.pem 
    ```

* `ziti-controller edge init`:

    ```
    $ ~/local/ziti/ziti-controller edge init local/etc/zrok.io/ziti-ctrl.yml
    ```

* start controller

* create and enroll edge router:

    ```
    $ ziti edge create edge-router router0 -o router0.jwt -t -a "public"
    New edge router router0 created with id: ZAbNbXUL6A
    Enrollment expires at 2022-08-29T21:56:37.418Z

    $ ziti-router enroll local/etc/zrok.io/ziti-router0.yml --jwt router0.jwt 
    [   3.561]    INFO edge/router/enroll.(*RestEnroller).Enroll: registration complete
    ```

* configure zrok frontend identity

    ```
    $ ziti edge create identity device -o ~/.zrok/proxy.jwt proxy
    New identity proxy created with id: -zbBF8eVb-
    Enrollment expires at 2022-08-10T18:46:16.641Z
    ```

    ```
    $ ziti edge enroll -j ~/.zrok/proxy.jwt -o ~/.zrok/proxy.json
    INFO    generating 4096 bit RSA key                  
    INFO    enrolled successfully. identity file written to: proxy.json
    ```

    ```
    $ ziti edge create erp --edge-router-roles "#all" --identity-roles @proxy
    ```
