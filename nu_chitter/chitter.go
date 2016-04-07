package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
    "strconv"
    "strings"
)

type user struct {
    uid string
    msgChan chan string    
    online bool
}

type usermsg struct {
    uid string
    msg string
}

// var msgChan = make(chan string)

func idManager(idAssignmentChan chan string)  {
    fmt.Println("idManager is up . . .")
    var uid uint64
    for uid = 0; ; uid++ {
        idAssignmentChan <- strconv.FormatUint(uid, 10)
    }   
}

func msgRouter(userManagementChan chan user, pubMsgChan chan usermsg)  {
    fmt.Println("msgRouter is up . . .")
    var user2chan = make(map[string]chan string)
    for {
        select {
            case oneUser := <-userManagementChan:
            // user online or offline
            if oneUser.online {
                fmt.Printf("chitter: user %s is online\n", oneUser.uid)
                user2chan[oneUser.uid] = oneUser.msgChan
            } else {
                fmt.Printf("chitter: user %s is offline\n", oneUser.uid)
                delete(user2chan, oneUser.uid)
            }
            case userMsg := <- pubMsgChan:
            ss := strings.SplitN(userMsg.msg, ":", 2)
            switch ss[0] {
            case "whoami":
                user2chan[userMsg.uid] <- "chitter: " + userMsg.uid + "\n"
            case "all":
                for uid := range user2chan {
                    user2chan[uid] <- userMsg.uid + ":" + ss[1]
                }
            default:
                if len(ss) == 2 {
                    c, ok := user2chan[ss[0]]
                    if ok {
                        c <- userMsg.uid + ":" + ss[1]
                    } else {
                        fmt.Printf("chitter: %s not a command or user id\n", ss[0])
                    }
                } else {
                    // len = 11
                    for uid := range user2chan {
                        user2chan[uid] <- userMsg.uid + ":" + userMsg.msg
                    }
                }
            }
            
        }
    }
}

func handleConnection(idAssignmentChan chan string, conn net.Conn, 
                      userManagment chan user, pubMsgChan chan usermsg)  {
    // private msgchan
    me := user {<-idAssignmentChan, make(chan string), true}
    // register in router service
    userManagment <- me
    sockChan := make(chan string)
    go func()  {
        // recv from sock
        buf := bufio.NewReader(conn)
        for {
            msg, err := buf.ReadString('\n')
            if err != nil {
                conn.Close()
                // signal user end of connection
                sockChan <- ""
                return
            }
            sockChan <- msg
        }
    } ()
    
    for {
        select {
            case msg := <-me.msgChan:
                conn.Write([]byte(msg))
            case msg := <-sockChan:
                // conn.Write([]byte(clientMsg))
                if msg == "" {
                    // end up this thread
                    me.online = false
                    userManagment <- me
                    return
                }
                pubMsgChan <- usermsg {me.uid, msg}
        }
    }
}

func main() {
    
    if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
    }
    port := os.Args[1]    
    // create sock and listen
    server, err := net.Listen("tcp", ":"+ port )
    if err != nil {
        fmt.Fprintf(os.Stderr, "chitter error: Can't bind to port %s\n", port)
        os.Exit(1)
    } else {
        fmt.Printf("chitter: start to listen on port %s . . .\n", port)
        // start and set up core services
        var idAssignmentChan = make(chan string)   
        var userManagmentChan = make(chan user)
        // all connection handle thread use this shared channel for sending userInput to router
        var pubMsgChan = make(chan usermsg)
        go idManager(idAssignmentChan)
        
        go msgRouter(userManagmentChan, pubMsgChan)
        for {
            conn, err := server.Accept()
            if err != nil {
                // channel will sync atomatically
                fmt.Fprintf(os.Stderr, "chitter error: Unable to accept a incoming connection\n")
            } else {
                go handleConnection(idAssignmentChan, conn, userManagmentChan, pubMsgChan)
            }
        }
    }
}