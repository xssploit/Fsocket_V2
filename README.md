# FSOCKET V2

**This is a simple websocket relay that forward websocket message from one client to another**

>run main.go

    go run main.go -p [port]
    eg: - go run main.go -p 8080

>run fsocket [linux executable]

    fsock -p [port]
    eg:- fsock -p 8080

>enable/disable website mode

    fsock -w=[true/false]
    eg:- fsock -w=false //disables website mode

>build (main.go)


    go build -ldflags="-s -w" main.go  //build
    upx --brute main  //lower size executable

>Run program (Simple)

    git cone https://github.com/xssploit/Fsocket_Server.git
    cd Fsocket_Server/
    chmod +x fsocket
    ./fsocket

> Testing

    Testing Websocket Chats with fsocket
    route : /testws
    status : available

    Testing Webrtc With fsocket
    route : /testwsrtc
    status : unavailable
