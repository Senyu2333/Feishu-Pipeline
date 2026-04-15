package service

import "time"

type HealthService struct {
	serviceName string
	version     string
}

func NewHealthService(serviceName string, version string) *HealthService {
	return &HealthService{
		serviceName: serviceName,
		version:     version,
	}
}

func (s *HealthService) Health() map[string]string {
	return map[string]string{
		"status":  "ok",
		"service": s.serviceName,
		"version": s.version,
		"now":     time.Now().UTC().Format(time.RFC3339),
	}
}
