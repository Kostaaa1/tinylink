package services

import (
	"github.com/Kostaaa1/tinylink/internal/tinylink/application/interfaces"
	"github.com/Kostaaa1/tinylink/internal/tinylink/domain/repositories"
)

type TinylinkService struct {
	tinylinkRepository repositories.Tinylink
}

func NewTinylinkService(
	tinylinkRepository repositories.Tinylink,
) interfaces.TinylinkService {
	return &TinylinkService{
		tinylinkRepository: tinylinkRepository,
	}
}
