package cli

type CliWallet struct {
	mnemonics string
	path      string
}

func NewCliWallet(configFile string) *CliWallet {
	return nil
}

func (c *CliWallet) toString() string {
	return "mnemonics: " + c.mnemonics + "\npath: " + c.path
}
