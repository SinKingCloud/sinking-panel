//go:build darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

func setTimezone(name string) error {
	database, err := openDatabase()
	if err != nil {
		return err
	}
	source, _, err := database.read(name)
	if err != nil {
		return err
	}
	return writeTimezone(name, source)
}
