package ip

type Info struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Districts string `json:"districts"`
	Isp       string `json:"isp"`
}
