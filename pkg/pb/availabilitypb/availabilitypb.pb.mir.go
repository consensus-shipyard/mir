package availabilitypb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_RequestCert) Unwrap() *RequestCert {
	return p.RequestCert
}

func (p *Event_NewCert) Unwrap() *NewCert {
	return p.NewCert
}

func (p *Event_VerifyCert) Unwrap() *VerifyCert {
	return p.VerifyCert
}

func (p *Event_CertVerified) Unwrap() *CertVerified {
	return p.CertVerified
}

func (p *Event_RequestTransactions) Unwrap() *RequestTransactions {
	return p.RequestTransactions
}

func (p *Event_ProvideTransactions) Unwrap() *ProvideTransactions {
	return p.ProvideTransactions
}
