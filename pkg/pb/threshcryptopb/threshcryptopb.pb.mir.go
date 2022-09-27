package threshcryptopb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_SignShare) Unwrap() *SignShare {
	return p.SignShare
}

func (p *Event_SignShareResult) Unwrap() *SignShareResult {
	return p.SignShareResult
}

func (p *Event_VerifyShare) Unwrap() *VerifyShare {
	return p.VerifyShare
}

func (p *Event_VerifyShareResult) Unwrap() *VerifyShareResult {
	return p.VerifyShareResult
}

func (p *Event_VerifyFull) Unwrap() *VerifyFull {
	return p.VerifyFull
}

func (p *Event_VerifyFullResult) Unwrap() *VerifyFullResult {
	return p.VerifyFullResult
}

func (p *Event_Recover) Unwrap() *Recover {
	return p.Recover
}

func (p *Event_RecoverResult) Unwrap() *RecoverResult {
	return p.RecoverResult
}
