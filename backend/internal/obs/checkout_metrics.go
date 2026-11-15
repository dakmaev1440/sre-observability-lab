package obs

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CheckoutPaid             = "paid"
	CheckoutBusinessDecline  = "business_decline"
	CheckoutTechnicalFailure = "technical_failure"
	CheckoutOutOfStock       = "out_of_stock"
)

const (
	ReasonNone                      = "none"
	ReasonCardDeclined              = "card_declined"
	ReasonPaymentServiceUnavailable = "payment_service_unavailable"
	ReasonOutOfStock                = "out_of_stock"
	ReasonOther                     = "other"
)

var checkoutSeries = [][2]string{
	{CheckoutPaid, ReasonNone},
	{CheckoutBusinessDecline, ReasonCardDeclined},
	{CheckoutTechnicalFailure, ReasonPaymentServiceUnavailable},
	{CheckoutOutOfStock, ReasonOutOfStock},
}

func newCheckoutAttempts() *prometheus.CounterVec {
	return newCounterVec("checkout_attempts_total",
		"Исходы оформления заказа (POST /api/orders). Ровно один инкремент на попытку.",
		"result", "failure_reason")
}

func newCheckoutDuration() *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "backend",
		Name:      "checkout_duration_seconds",
		Help:      "Длительность оформления заказа по классу исхода.",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 10},
	}, []string{"result"})
	registerBM(h)
	return h
}

func (b *BusinessMetrics) initCheckoutSeries() {
	for _, s := range checkoutSeries {
		b.CheckoutAttempts.WithLabelValues(s[0], s[1])
		b.CheckoutDuration.WithLabelValues(s[0])
	}
}

func normalizeReason(result, reason string) string {
	for _, s := range checkoutSeries {
		if s[0] == result && s[1] == reason {
			return reason
		}
	}
	return ReasonOther
}

func (b *BusinessMetrics) RecordCheckout(result, reason string, d time.Duration) {
	if b == nil || b.CheckoutAttempts == nil || b.CheckoutDuration == nil {
		return
	}
	b.CheckoutAttempts.WithLabelValues(result, normalizeReason(result, reason)).Inc()
	b.CheckoutDuration.WithLabelValues(result).Observe(d.Seconds())
}