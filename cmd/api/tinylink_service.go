package main

type TinylinkService struct {
	TinylinkRepository TinylinkRepository
}

func NewTinylinkService(tlRepo TinylinkRepository) *TinylinkService {
	return &TinylinkService{TinylinkRepository: tlRepo}
}
