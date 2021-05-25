package main

import (
	"fmt"
	"net"
	"os"

	"github.com/btcsuite/btcd/wire"
	"github.com/millken/golog"
)

const servAddr = "123.115.223.138:8333"

func main() {
	fh := &golog.FileHandler{
		Output: os.Stdout,
	}
	fh.SetLevel(golog.DebugLevel)
	fh.SetFormatter(&golog.TextFormatter{
		EnableCaller: true,
	})
	logger := golog.NewLogger()
	logger.AddHandler(fh)
	// Create version message data.
	lastBlock := int32(0)
	tcpAddrMe := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8333}
	me := wire.NewNetAddress(tcpAddrMe, wire.SFNodeNetwork)
	tcpAddrYou := &net.TCPAddr{IP: net.ParseIP("123.115.223.138"), Port: 9333}
	you := wire.NewNetAddress(tcpAddrYou, wire.SFNodeNetwork)
	nonce, err := wire.RandomUint64()
	if err != nil {
		fmt.Printf("RandomUint64: error generating nonce: %v", err)
	}

	// Ensure we get the correct data back out.
	msg := wire.NewMsgVersion(me, you, nonce, lastBlock)
	msg.AddService(wire.SFNodeNetwork)
	conn, err := net.Dial("tcp", servAddr)
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	err = wire.WriteMessage(conn, msg, wire.ProtocolVersion, wire.MainNet)

	if err != nil {
		logger.Fatal(err)
	}
	for {
		remoteMsg, _, err := wire.ReadMessage(conn, wire.ProtocolVersion, wire.MainNet)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debug(remoteMsg.Command())
		switch remoteMsg.Command() {
		case wire.CmdPing:
			msg := remoteMsg.(*wire.MsgPing)
			logger.WithFields(golog.Fields{"nonce": msg.Nonce}).Info("receive ping")
			wire.WriteMessage(conn, wire.NewMsgPong(msg.Nonce), wire.ProtocolVersion, wire.MainNet)
		case wire.CmdVersion:
			msg := remoteMsg.(*wire.MsgVersion)
			logger.WithFields(golog.Fields{"blockHeight": msg.LastBlock}).Info(msg.ProtocolVersion)

		case wire.CmdVerAck:
			wire.WriteMessage(conn, wire.NewMsgVerAck(), wire.ProtocolVersion, wire.MainNet)
		case wire.CmdInv:
			// msg := remoteMsg.(*wire.MsgInv)
			// for _, inv := range msg.InvList {
			// 	logger.WithFields(golog.Fields{"hash": inv.Hash.String()}).Info(inv.Type.String())
			// }
		}

	}
}
