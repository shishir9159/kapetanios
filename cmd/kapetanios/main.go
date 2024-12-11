package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/swagger"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	// TODO: or should it be klovercloud with additional service accounts?
	certRenewalNamespace  = "default"
	minorUpgradeNamespace = "default"
)

func certRenewal(c *fiber.Ctx) error {

	go Cert(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func cleanup(c *fiber.Ctx) error {

	go Cleanup(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func expiration(c *fiber.Ctx) error {

	go Expiration(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func minorUpgrade(c *fiber.Ctx) error {

	go MinorUpgradeFirstRun(minorUpgradeNamespace)

	//go MinorUpgradeFirstRun(minorUpgradeNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func collback(w http.ResponseWriter, r *http.Request) {

	go Rollback(certRenewalNamespace)

	// Upgrade the connection to WebTransport
	w.Header().Set("Content-Type", "text/plain")
	r.
	if !http3.CheckIsExtendedConnect(r) {
		http.Error(w, "Expected a WebTransport connect", http.StatusBadRequest)
		return
	}

	log.Println("New WebTransport connection")
	session, ok := w.(http3.WebTransportSession)
	if !ok {
		http.Error(w, "Failed to create WebTransport session", http.StatusInternalServerError)
		return
	}

	// Start streaming logs to the client
	go streamLogs(session)

	// Handle incoming streams for acknowledgments
	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Println("Error accepting stream:", err)
			break
		}
		go handleAcknowledgments(stream)
	}

}

func rollback(c *fiber.Ctx) error {

	go Rollback(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func sanityChecking(c *fiber.Ctx) error {

	c.Accepts(`shuttle="launched"`)
	c.Status(http.StatusOK)

	return nil
}

func shuttleLaunched(c *fiber.Ctx) error {

	c.Accepts(`sanity="checked"`)
	c.Status(http.StatusOK)

	return nil
}

func setupRoutes(app *fiber.App) {

	app.Get("/readyz", sanityChecking)
	app.Get("/livez", shuttleLaunched)
	app.Get("/renewal", certRenewal)
	app.Get("/cleanup", cleanup)
	app.Get("/expiration", expiration)
	app.Get("/minor-upgrade", minorUpgrade)
	app.Get("/rollback", rollback)
	app.Get("/swagger", swagger.HandlerDefault)

	//api.Group("")
	//	.SetupRoutes(cert)
	//	.SetupRoutes(minorUpgrade)
}

// decide if you need a separate router folder or not. more like you are gonna need it

func SetupGroupRoutes(router fiber.Router) {

}

func tmpMain() {
		quicConfig := &quic.Config{
			GetConfigForClient:             nil,
			Versions:                       nil,
			HandshakeIdleTimeout:           0,
			MaxIdleTimeout:                 0,
			TokenStore:                     nil,
			InitialStreamReceiveWindow:     0,
			MaxStreamReceiveWindow:         0,
			InitialConnectionReceiveWindow: 0,
			MaxConnectionReceiveWindow:     0,
			AllowConnectionWindowIncrease:  nil,
			MaxIncomingStreams:             0,
			MaxIncomingUniStreams:          0,
			KeepAlivePeriod:                0,
			InitialPacketSize:              0,
			DisablePathMTUDiscovery:        false,
			Allow0RTT:                      false,
			EnableDatagrams:                false,
			Tracer:                         nil,
		}

		// Create a WebTransport server
		transportServer := &webtransport.Server{
		H3: http3.Server{
			Addr: "",
			Port: 0,
			TLSConfig: &tls.Config{
				Rand:                                nil,
				Time:                                nil,
				Certificates:                        nil,
				NameToCertificate:                   nil,
				GetCertificate:                      nil,
				GetClientCertificate:                nil,
				GetConfigForClient:                  nil,
				VerifyPeerCertificate:               nil,
				VerifyConnection:                    nil,
				RootCAs:                             &x509.CertPool{},
				NextProtos:                          nil,
				ServerName:                          "",
				ClientAuth:                          0,
				ClientCAs:                           &x509.CertPool{},
				InsecureSkipVerify:                  false,
				CipherSuites:                        nil,
				PreferServerCipherSuites:            false,
				SessionTicketsDisabled:              false,
				SessionTicketKey:                    [32]byte{},
				ClientSessionCache:                  nil,
				UnwrapSession:                       nil,
				WrapSession:                         nil,
				MinVersion:                          0,
				MaxVersion:                          0,
				CurvePreferences:                    nil,
				DynamicRecordSizingDisabled:         false,
				Renegotiation:                       0,
				KeyLogWriter:                        nil,
				EncryptedClientHelloConfigList:      nil,
				EncryptedClientHelloRejectionVerify: nil,
			},
			QUICConfig: quicConfig,
			Handler:            nil,
			EnableDatagrams:    false,
			MaxHeaderBytes:     0,
			AdditionalSettings: nil,
			StreamHijacker:     nil,
			UniStreamHijacker:  nil,
			IdleTimeout:        0,
			ConnContext:        nil,
			//Logger:             &slog.Logger{},
		},
		ReorderingTimeout: 0,
		CheckOrigin:       nil,
	}

	// Create a new HTTP endpoint /webtransport.
	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		sess, err := transportServer.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}
		// Handle the session. Here goes the application logic.

		stream, err := sess.OpenStream()
		if err != nil {
			log.Printf("opening stream failed: %s", err)
			return
		}

		p := []bytes.Buffer()

		read, err := stream.Read(p)
		if err != nil {
			return
		}


	})

	err := transportServer.ListenAndServeTLS("cert.pem", "ca.key")
	if err != nil {
		return
	}





	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := wtServer.ServeHTTP(w, r); err != nil {
			fmt.Printf("ServeHTTP error: %v\n", err)
		}
	})

	go func()
	for {

	session, err := wtServer.Accept()

	if err != nil {
	fmt.Printf("Failed to accept session: %v\n", err)
	continue
	}
	fmt.Println("New WebTransport session established")

	go handleSession(session)
	}
}()

 fmt.Println("Starting WebTransport server on :4433")
err := http3.ListenAndServeQUIC(":4433", "cert.pem", "key.pem", nil)
if err != nil {
	fmt.Printf("Failed to start server: %v\n", err)
	os.Exit(1)
}
}

func handleSession(session *webtransport.Session) {
	defer session.Close()
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			fmt.Printf("Error accepting stream: %v\n", err)
			break
		}
		go handleStream(stream)
	}
}

func handleStream(stream webtransport.Stream) {
	defer stream.Close()
	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			fmt.Printf("Stream closed: %v\n", err)
			break
		}
		fmt.Printf("Received: %s\n", buf[:n])
	}
}


func main() {

	s := webtransport.Server{
		H3: http3.Server{
			Addr:               ":443",
			Port:               0,
			TLSConfig:          &tls.Config{

			}, // use your tls.Config here
			QUICConfig:         nil,
			Handler:            http.HandlerFunc(collback),
			EnableDatagrams:    false,
			MaxHeaderBytes:     0,
			AdditionalSettings: nil,
			StreamHijacker:     nil,
			UniStreamHijacker:  nil,
			IdleTimeout:        0,
			ConnContext:        nil,
			Logger:             nil,
		},
		ReorderingTimeout: 0,
		CheckOrigin:       nil,
	}

	// Create a new HTTP endpoint /webtransport.
	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.Upgrade(w, r)
		if err != nil {
			//log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}
		// Handle the session. Here goes the application logic.
	})

	err := s.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		return
	}

	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1234})
	// ... error handling
	tr := quic.Transport{
		Conn: udpConn,
	}
	ln, err := tr.Listen(tlsConf, quicConf)
	// ... error handling
	for {
		conn, err := ln.Accept()
		// ... error handling
		// handle the connection, usually in a new Go routine
	}

	app := fiber.New()

	//logger, err := zap.NewProduction()

	//if err != nil {
	//}

	//app.Use(recover.New())
	//app.Use(compress.New())

	//app.Use(fiberzap.New(fiberzap.Config{
	//	Logger: logger,
	//}))

	// TODO:
	//  sync.pool for caching mechanism

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	// TODO:
	//  Controller Definition need to be moved with the
	//  initial Setup and making sure there exists only one

	// setup routes
	setupRoutes(app)

	//app.Use("/livez", healthcheck.New())

	//app.Use(healthcheck.New(healthcheck.Config{
	//	LivenessProbe: func(c *fiber.Ctx) bool {
	//		return true
	//	},
	//	LivenessEndpoint: "/livez",
	//	ReadinessProbe: func(c *fiber.Ctx) bool {
	//		return true
	//	},
	//	ReadinessEndpoint: "/readyz",
	//}))

	Prerequisites(minorUpgradeNamespace)

	// TODO: prevent duplicate lighthouse instances
	// TODO: websocket
	err := app.Listen(":80")

	if err != nil {
		return
	}

}
