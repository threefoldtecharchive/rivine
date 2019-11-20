package tftschema

import (
	"github.com/threefoldtech/zos/pkg/schema"
	"time"
)

func DateFromTimestamp(secs int64) schema.Date {
	return schema.Date{
		Time: time.Unix(secs, 0),
	}
}
