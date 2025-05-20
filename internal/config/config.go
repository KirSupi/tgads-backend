package config

import (
	"github.com/timmbarton/layout/components/httpserver"
	"github.com/timmbarton/layout/components/postgresconn"
	"github.com/timmbarton/layout/components/tracingconn"

	"backend/internal/usecase"
)

type Config struct {
	HTTP     httpserver.Config
	Postgres postgresconn.Config
	Tracing  tracingconn.Config
	UseCase  usecase.Config
}
