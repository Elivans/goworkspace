package db

import (
	"fmt"
)

type Memu1 struct {
	Name   string
	Url    string
	Active string
}

type Memu0 struct {
	Icon    string
	Name    string
	Url     string
	Active  string
	SubMemu []Memu1
}

func GetMemu(user string, urlNow string) (memu0s []Memu0, memuName string, err error) {
	fmt.Println(urlNow)
	//加载菜单,todo
	var m0 Memu0
	m0.Icon = "menu-icon fa fa-desktop"
	m0.Name = "主菜单1"
	m0.Active = "Y"

	var m1 Memu1
	m1.Name = "任务信息"
	m1.Url = "workinfo"
	m0.SubMemu = append(m0.SubMemu, m1)
	var m2 Memu1
	m2.Name = "子菜单2"
	m2.Url = "#2"
	m2.Active = "Y"
	m0.SubMemu = append(m0.SubMemu, m2)

	memuName = "子菜单2" //选中的菜单名

	memu0s = append(memu0s, m0)
	memu0s = append(memu0s, m0)
	return
}
