package main

import (
	"github.com/ant0ine/go-urlrouter"
	"github.com/wurkhappy/WH-Email/handlers"
)

var router urlrouter.Router = urlrouter.Router{
	Routes: []urlrouter.Route{
		urlrouter.Route{
			PathExp: "agreement.submitted",
			Dest:    handlers.AgreementSubmitted,
		},
		urlrouter.Route{
			PathExp: "agreement.accepted",
			Dest:    handlers.AgreementAccept,
		},
		urlrouter.Route{
			PathExp: "agreement.rejected",
			Dest:    handlers.AgreementReject,
		},
		urlrouter.Route{
			PathExp: "payment.submitted",
			Dest:    handlers.PaymentRequest,
		},
		urlrouter.Route{
			PathExp: "payment.accepted",
			Dest:    handlers.PaymentAccepted,
		},
		urlrouter.Route{
			PathExp: "payment.rejected",
			Dest:    handlers.PaymentReject,
		},
	},
}
