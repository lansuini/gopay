package model

type PayType struct {
	D0Settlement  string `json:"D0Settlement"`
	Gcash         string `json:"gcash"`
	Grabpay       string `json:"grabpay"`
	Seven11Direct string `json:"711_direct"`
	Da5           string `json:"da5"`
	Qr            string `json:"qr"`
	Payngo        string `json:"payngo"`
	Posible       string `json:"posible"`
	RLNT          string `json:"RLNT"`
	RDS           string `json:"RDS"`
	UBPB          string `json:"UBPB"`
	RCBC          string `json:"RCBC"`
	SMR           string `json:"SMR"`
	LBC           string `json:"LBC"`
	Ussc          string `json:"ussc"`
	PLWN          string `json:"PLWN"`
	Partnerpay    string `json:"partnerpay"`
	RDP           string `json:"RDP"`
	BPIA          string `json:"BPIA"`
	ECPY          string `json:"ECPY"`
	CEBL          string `json:"CEBL"`
}
