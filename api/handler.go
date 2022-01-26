package api

import (
	"github.com/rovergulf/chain/database"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.SugaredLogger
	db     database.Backend
}
