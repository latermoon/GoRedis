package main

// 伪代码
import (
	"fmt"
)

func main() {

}

func sync() {

}

func (server *GoRedisServer) OnSYNC(session *Session, cmd *Command) (reply *Reply) {
}

// Master:
// SYNC_BULK [SEQ] [CMD]

// Slave:
// SYNC_BULK_FIN [SEQ]
func (server *GoRedisServer) OnSYNC_FIN(session *Session, cmd *Command) (reply *Reply) {
	// <- SYNC_FIN [seq]
	// db.CleanBulks(success)
	// session.WriteCommand(SYNC_BULK ...)
	// session.WriteCommand(SYNC_BULK_FIN)
}

// slave
func (server *GoRedisServer) OnSYNC_BULK(session *Session, cmd *Command) (reply *Reply) {
	// <- SYNC_BULK cmd
	// bulks = append(bulks, cmd)
	// NO_REPLY
}

// -> SYNC_FIN [seq]
func (server *GoRedisServer) OnSYNC_BULK_FIN(session *Session, cmd *Command) (reply *Reply) {
	// <- SYNC_BULK_SEQ [seq]
	// db.Write(bulks)
	// -> SYNC_FIN [seq]
}
