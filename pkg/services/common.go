package services

import "sync"

type (
	Service interface {
		Run() error
		Close(group *sync.WaitGroup)
	}

	ServiceBuilder func(config *ProxyConfig, index int) Service
)
