package ops

func InitBasic() {
	//http server
	OPS_Server = new(OPSServer)
	//warning related
	go NetSpeed()
	go BindNum()
}
