package db

type PrvMcode struct {
	Mcode string
	Mname string
}

func GetParam(gocde string) (mcodes []PrvMcode, err error) {
	//取参数,todo
	var m1 PrvMcode
	m1.Mcode = "0"
	m1.Mname = "正常"

	mcodes = append(mcodes, m1)
	var m2 PrvMcode
	m2.Mcode = "1"
	m2.Mname = "不正常"

	mcodes = append(mcodes, m2)

	return
}
