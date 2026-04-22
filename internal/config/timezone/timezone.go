package timezone

import "time"

var ParaguayTZ = func() *time.Location {
	loc, err := time.LoadLocation("America/Asuncion")
	if err != nil {
		panic(err)
	}
	return loc
}()
