package datagen

import (
	"fmt"
	"math/rand"

	model "mcolomerc/synth-payment-producer/pkg/avro"
	"time"

	"github.com/go-faker/faker/v4"
)

type Datagen struct {
	Sources      []string
	Destinations []string
}

func NewDatagen(sourcesNum int, destinationsNum int) Datagen {
	var sources []string
	var destinations []string
	for i := 0; i < sourcesNum; i++ {
		sources = append(sources, fmt.Sprintf("bank-%v", i))
	}
	for i := 0; i < destinationsNum; i++ {
		destinations = append(destinations, fmt.Sprintf("bank-%v", i))
	}
	return Datagen{
		Sources:      sources,
		Destinations: destinations,
	}
}

func (d *Datagen) GeneratePayment() model.Payment {
	// Build Payment
	max := 9999.0
	min := 0.1
	// Get random source
	source := d.Sources[rand.Intn(len(d.Sources))]
	destination := d.Destinations[rand.Intn(len(d.Destinations))]
	return model.Payment{
		Id:          faker.UUIDDigit(),
		Ts:          time.Now().UnixNano() / 1e6,
		Destination: destination,
		Source:      source,
		Currency:    faker.Currency(),
		Amount:      min + rand.Float64()*(max-min),
		Status:      Initiated.String(),
	}
}
