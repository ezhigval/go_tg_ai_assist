package credits

import (
	"math"
	"time"
)

// CalcAnnuitySchedule — считает аннуитетный график.
func CalcAnnuitySchedule(
	principal float64,
	rateAnnual float64,
	months int,
	startDate time.Time,
	paymentDay int,
) (monthly float64, schedule []Payment) {

	r := rateAnnual / 12.0 / 100.0
	if r <= 0 {
		// без процентов — просто делим
		monthly = principal / float64(months)
	} else {
		// A = P * r * (1+r)^n / ((1+r)^n - 1)
		pow := math.Pow(1+r, float64(months))
		monthly = principal * r * pow / (pow - 1)
	}
	monthly = math.Round(monthly*100) / 100

	remaining := principal
	date := normaliseDate(startDate, paymentDay)

	for i := 0; i < months; i++ {
		interest := remaining * (rateAnnual / 12.0 / 100.0)
		interest = math.Round(interest*100) / 100

		principalPart := monthly - interest
		principalPart = math.Round(principalPart*100) / 100

		if principalPart > remaining {
			principalPart = remaining
		}
		remaining -= principalPart
		if remaining < 0 {
			remaining = 0
		}

		schedule = append(schedule, Payment{
			DueDate:   date,
			Principal: principalPart,
			Interest:  interest,
			Total:     monthly,
		})

		date = date.AddDate(0, 1, 0)
		date = normaliseDate(date, paymentDay)
	}

	return monthly, schedule
}

// CalcDiffSchedule — дифференцированный график.
func CalcDiffSchedule(
	principal float64,
	rateAnnual float64,
	months int,
	startDate time.Time,
	paymentDay int,
) (schedule []Payment) {

	base := principal / float64(months)
	base = math.Round(base*100) / 100

	remaining := principal
	date := normaliseDate(startDate, paymentDay)

	for i := 0; i < months; i++ {
		interest := remaining * (rateAnnual / 12.0 / 100.0)
		interest = math.Round(interest*100) / 100

		total := base + interest
		total = math.Round(total*100) / 100

		if base > remaining {
			base = remaining
		}
		remaining -= base
		if remaining < 0 {
			remaining = 0
		}

		schedule = append(schedule, Payment{
			DueDate:   date,
			Principal: base,
			Interest:  interest,
			Total:     total,
		})

		date = date.AddDate(0, 1, 0)
		date = normaliseDate(date, paymentDay)
	}
	return schedule
}

func normaliseDate(base time.Time, day int) time.Time {
	if day <= 0 {
		day = base.Day()
	}
	year, month, _ := base.Date()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, base.Location()).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, base.Location())
}
