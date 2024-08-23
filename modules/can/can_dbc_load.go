package can

func (mod *CANModule) dbcLoad(name string) error {
	// load as file
	return mod.dbc.LoadFile(mod, name)
}
