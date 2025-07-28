package mq

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/consumer"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
)

var WireSet = wire.NewSet(
	consumer.WireSet,
	producer.WireSet,
)
