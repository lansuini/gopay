package menu

type GmMenu []struct {
	N  string   `json:"n"`
	U  string   `json:"u,omitempty"`
	Pu []string `json:"pu,omitempty"`
	ID int      `json:"id"`
	C  []struct {
		N  string   `json:"n"`
		U  string   `json:"u"`
		Pu []string `json:"pu"`
		ID int      `json:"id"`
	} `json:"c,omitempty"`
}

type MerchantMenu []struct {
	N  string `json:"n"`
	U  string `json:"u,omitempty"`
	ID int    `json:"id"`
	C  []struct {
		N  string `json:"n"`
		U  string `json:"u"`
		ID int    `json:"id"`
	} `json:"c,omitempty"`
}
