package cache

import (
	"fmt"

	"github.com/olegshs/go-tools/cache/storage"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/logs"
)

type Log struct {
	storage  StorageInterface
	channel  *logs.LogChannel
	interval *helpers.Interval
}

func NewLog(storage StorageInterface, conf storage.ConfigLog) *Log {
	log := new(Log)
	log.storage = storage
	log.channel = logs.Channel(conf.Channel)
	log.interval = helpers.NewInterval(conf.Interval, func() {
		log.printStats()
	})

	return log
}

func (log *Log) Start() {
	log.interval.Start()
}

func (log *Log) Stop() {
	log.interval.Stop()
}

func (log *Log) printStats() {
	hits := log.storage.Hits()
	misses := log.storage.Misses()

	log.channel.Debug("cache:", fmt.Sprintf(
		"hits=%d, misses=%d",
		hits, misses,
	))
}
