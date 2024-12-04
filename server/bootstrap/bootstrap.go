package bootstrap

func Load() {
	LoadInstall()
	LoadConf()
	LoadLog()
	LoadCache()
	LoadDatabase()
	LoadSign()
}
