package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"net"
	"qrpay-wpp/configs"
	"qrpay-wpp/internal/api/system"
)

type Server struct {
	settingsPath string

	db      *pgxpool.Pool
	context context.Context

	repos    *repositories
	handlers *handlers
	services *services
}

func (s *Server) createDatabaseIfNotExists(err error) error {
	err = errors.Unwrap(err)
	if er, ok := err.(*pgconn.PgError); ok {
		if er.Code == "3D000" {
			c := configs.Get().Database
			url := fmt.Sprintf("postgres://%s:%s@%s:%d/?sslmode=disable", c.Username, c.Password, c.Host, c.Port)
			db, err := pgxpool.New(s.context, url)
			if err != nil {
				return fmt.Errorf("unable to connect to database: %v", err)
			}
			defer db.Close()
			query := fmt.Sprintf("CREATE DATABASE %s", c.Name)
			_, err = db.Exec(s.context, query)
			if err != nil {
				return fmt.Errorf("unable to create database: %v", err)
			}
		}
	}
	return nil
}

func (s *Server) startDatabase() error {
	// Create a new context
	s.context = context.Background()

	// Create a new connection pool
	c := configs.Get().Database
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Name)
	db, err := pgxpool.New(s.context, url)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	// Ping the database to check if it's still alive.
	err = db.Ping(s.context)
	if err != nil {
		err = s.createDatabaseIfNotExists(err)
		if err == nil {
			return s.startDatabase()
		}
		return fmt.Errorf("unable to ping database: %v", err)
	}

	// Set the database connection pool to the Server struct
	s.db = db
	return nil
}

func (s *Server) initializeAPI() error {

	// Create Repositories
	err := s.createRepositories()
	if err != nil {
		return err
	}

	wppSystem, err := system.New()
	if err != nil {
		return err
	}

	// Create Services
	s.createServices(wppSystem)

	// Create Handlers
	s.createHandlers()

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the Services
	s.registerServices(grpcServer)

	// Create a new TCP listener
	address := fmt.Sprintf(":%d", configs.Get().Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Start the gRPC server
	fmt.Printf("gRPC server listening at %s\n", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}

func New(settingsPath string) *Server {
	return &Server{
		settingsPath: settingsPath,
		repos:        &repositories{},
		handlers:     &handlers{},
		services:     &services{},
	}
}

func (s *Server) Start() error {
	err := configs.Load(s.settingsPath)
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}
	err = s.startDatabase()
	if err != nil {
		return fmt.Errorf("could not start database: %w", err)
	}
	err = s.initializeAPI()
	if err != nil {
		return fmt.Errorf("could not start api: %w", err)
	}
	return nil
}
