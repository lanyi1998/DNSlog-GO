package Core

var Config = struct {
	HTTP struct {
		Port           string
		Token          string
		ConsoleDisable bool
	}
	Dns struct{
		Domain string
	}
}{}
