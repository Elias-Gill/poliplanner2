package timezone

import (
	"time"
	// Embed the IANA timezone database so ParaguayTZ never panics on minimal
	// container images that lack system tzdata.
	_ "time/tzdata"
)

var ParaguayTZ = func() *time.Location {
	loc, err := time.LoadLocation("America/Asuncion")
	if err != nil {
		panic(err)
	}
	return loc
}()
