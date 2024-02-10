package datagen

import (
	"math/rand"
	"strings"

	model "mcolomerc/synth-payment-producer/pkg/avro"
	"time"

	"github.com/go-faker/faker/v4"
)

type Datagen struct {
	Sources      []model.Bank
	Destinations []model.Bank
}

func NewDatagen(sourcesNum int, destinationsNum int) Datagen {
	var sources []model.Bank
	var destinations []model.Bank
	for i := 0; i < sourcesNum; i++ {
		sources = append(sources, generateBank())
	}
	for i := 0; i < destinationsNum; i++ {
		destinations = append(destinations, generateBank())
	}
	return Datagen{
		Sources:      sources,
		Destinations: destinations,
	}
}

func (d *Datagen) GeneratePayment() model.Payment {
	// Build Payment
	max := 99999.0
	min := 0.1
	// Get random source
	source := d.Sources[rand.Intn(len(d.Sources))]
	destination := d.Destinations[rand.Intn(len(d.Destinations))]
	return model.Payment{
		Id:          faker.UUIDDigit(),
		Ts:          time.Now().UnixNano() / 1e6,
		Destination: destination.Id,
		Source:      source.Id,
		Currency:    faker.Currency(),
		Amount:      min + rand.Float64()*(max-min),
		Status:      Initiated.String(),
	}
}

func (d *Datagen) GetBanks() []model.Bank {
	return append(d.Sources, d.Destinations...)
}
func generateBank() model.Bank {
	bankCode := strings.ToLower(faker.FirstName())
	return model.Bank{
		Id:         faker.UUIDDigit(),
		Name:       bankCode + "-bank",
		Country:    faker.Currency(),
		Email:      faker.Email(),
		Website:    faker.URL(),
		BankCode:   faker.FirstName(),
		Bic:        faker.UUIDDigit(),
		Branch:     faker.UUIDDigit(),
		Updated_ts: time.Now().Format(time.RFC3339),
		Created_ts: time.Now().Format(time.RFC3339),
		Version:    0,
	}
}
