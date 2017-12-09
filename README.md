# demo-controller

This is a demo-controller application.

To deploy this application, execute the following commands:

  1. Clone repo:

    ```
    $ git clone https://github.com/damianjaniszewski/demo-controller
    $ cd demo-controller
    ```

  2. Install Godep package manager (git required to complete):

    ```
    $ go get github.com/tools/godep

    ```

  3. Create Godep package manager files:

    ```
    $ godep save
    ```

  3. Deploy to HPE Stackato PaaS

    ```
    $ stackato push --stack cflinuxfs2 -n
    ```
