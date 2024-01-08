package datagen

import (
	"strings"
)

type Status string

const (
	Initiated Status = "Initiated"
	Completed Status = "Completed"
	Failed    Status = "Failed"
	Canceled  Status = "Canceled"
	Validated Status = "Validated"
	Accounted Status = "Accounted"
	Rejected  Status = "Rejected"
)

var statusMap = map[string]Status{
	"initiated": Initiated,
	"completed": Completed,
	"failed":    Failed,
	"canceled":  Canceled,
	"validated": Validated,
	"accounted": Accounted,
	"rejected":  Rejected,
}

func (s Status) String() string {
	return string(s)
}

func GetStatus(str string) Status {
	c, _ := statusMap[strings.ToLower(str)]
	return c
}

func GetStatusList() []Status {
	return []Status{Initiated, Completed, Failed, Canceled, Validated, Accounted, Rejected}
}
