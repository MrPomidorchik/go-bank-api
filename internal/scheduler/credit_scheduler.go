package scheduler

import (
	"log"
	"time"

	"bank-api/internal/service"
)

type CreditScheduler struct {
	creditService service.CreditService
	interval      time.Duration
}

func NewCreditScheduler(creditService service.CreditService, interval time.Duration) *CreditScheduler {
	return &CreditScheduler{
		creditService: creditService,
		interval:      interval,
	}
}

func (s *CreditScheduler) Start() {
	ticker := time.NewTicker(s.interval)

	go func() {
		for {
			<-ticker.C

			result, err := s.creditService.ProcessDuePayments()
			if err != nil {
				log.Printf("credit scheduler error: %v", err)
				continue
			}

			log.Printf(
				"credit scheduler processed: total=%d paid=%d overdue=%d",
				result.ProcessedPayments,
				result.PaidPayments,
				result.OverduePayments,
			)
		}
	}()
}
