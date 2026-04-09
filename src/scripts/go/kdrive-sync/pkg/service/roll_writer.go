package service

import "kdrive-sync/pkg/domain"

// RollWriter persists a Roll as an Astro content file.
type RollWriter interface {
	WriteRoll(slug string, roll *domain.Roll) error
}
