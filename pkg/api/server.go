package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/betterde/cdns/api/routes"
	"github.com/betterde/cdns/config"
	"github.com/betterde/cdns/internal/journal"
	"github.com/betterde/cdns/internal/response"
	"github.com/betterde/cdns/pkg/challenge"
	"github.com/betterde/cdns/pkg/dns"
	"github.com/caddyserver/certmagic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
)

var ServerInstance *Server

type Server struct {
	Engine *fiber.App
}

func InitServer(name, version string) {
	ServerInstance = &Server{
		Engine: fiber.New(fiber.Config{
			AppName:       name,
			ServerHeader:  fmt.Sprintf("%s %s", name, version),
			CaseSensitive: true,
			// Override default error handler
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				// Status code defaults to 500
				code := fiber.StatusInternalServerError

				// Retrieve the custom status code if it's a fiber.*Error
				var e *fiber.Error
				if errors.As(err, &e) {
					code = e.Code
				}

				if err != nil {
					if code >= fiber.StatusInternalServerError {
						journal.Logger.Sugar().Errorw("Analysis server runtime error:", zap.Error(err))
					}

					// In case the SendFile fails
					return ctx.Status(code).JSON(response.Send(code, err.Error(), err))
				}

				return nil
			},
		}),
	}

	routes.RegisterRoutes(ServerInstance.Engine)
}

func (s *Server) Run(verbose bool, errChan chan error) {
	ServerInstance.Engine.Use(cors.New())

	if verbose {
		ServerInstance.Engine.Use(logger.New())
	}
	ServerInstance.Engine.Use(pprof.New())
	ServerInstance.Engine.Use(recover.New())
	ServerInstance.Engine.Use(requestid.New())

	go func() {
		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS11,
			MaxVersion: tls.VersionTLS13,
		}

		switch config.Conf.HTTP.TLS.Mode {
		case config.TLSModeACME:
			provider := challenge.NewChallengeProvider(dns.Servers)
			storage := certmagic.FileStorage{Path: config.Conf.Providers.ACME.Storage}

			certmagic.DefaultACME.DNS01Solver = &provider
			certmagic.DefaultACME.Agreed = true
			certmagic.DefaultACME.CA = config.Conf.Providers.ACME.Server
			certmagic.DefaultACME.Email = config.Conf.Providers.ACME.Email
			certmagic.DefaultACME.Logger = journal.Logger

			magicConf := &certmagic.Config{}
			magicConf.OCSP = certmagic.OCSPConfig{
				DisableStapling: true,
			}
			magicConf.Logger = journal.Logger
			magicConf.Storage = &storage
			magicConf.DefaultServerName = config.Conf.HTTP.Domain

			magicCache := certmagic.NewCache(certmagic.CacheOptions{
				Logger: journal.Logger,
				GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
					return magicConf, nil
				},
			})

			magic := certmagic.New(magicCache, *magicConf)

			err := magic.ManageAsync(context.Background(), []string{config.Conf.HTTP.Domain})
			if err != nil {
				errChan <- err
				return
			}

			tlsConf.GetCertificate = magic.GetCertificate
			tlsConf.NextProtos = []string{"http/1.1", "acme-tls/1"}

			// Create custom listener
			ln, err := tls.Listen("tcp", config.Conf.HTTP.Listen, tlsConf)
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}

			err = ServerInstance.Engine.Listener(ln)
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}
			break
		case config.TLSModeFile:
			cert, err := tls.LoadX509KeyPair("certs/ssl.cert", "certs/ssl.key")
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}

			// Create custom listener
			ln, err := tls.Listen("tcp", config.Conf.HTTP.Listen, &tls.Config{Certificates: []tls.Certificate{cert}})
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}

			err = ServerInstance.Engine.Listener(ln)
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}
			break
		default:
			err := ServerInstance.Engine.Listen(config.Conf.HTTP.Listen)
			if err != nil {
				journal.Logger.Sugar().Panicw("Failed to start cdns server:", err)
			}
		}
	}()
}
