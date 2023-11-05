package QrB

import . "github.com/x-ream/sqlxb"

type QrB struct {
	CondBuilder
	rs []Bb

	po Po
}

func Of(po Po) *QrB {
	var qrb = new(QrB)
	qrb.po = po
	return qrb
}

func (qrb *QrB) Cond(cond func(cb *CondBuilder)) *QrB {
	cond(&qrb.CondBuilder)
	return qrb
}

func (qrb *QrB) Eq(k string, v interface{}) *QrB {
	qrb.CondBuilder.Eq(k, v)
	return qrb
}

func (qrb *QrB) Build() *Qr {

	qr := Qr{
		Rs: qrb.rs,
		Cs: qrb.CondBuilder.Bbs,
	}

	return &qr
}
